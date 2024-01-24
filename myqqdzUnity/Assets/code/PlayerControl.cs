using System.Collections.Generic;
using UnityEngine;
using System.Net.Sockets;
using Google.Protobuf;
using logicProto;
using UnityEngine.UI;
using System;
using System.IO;

public class PlayerControl : MonoBehaviour {
	// 连接信息
	public Echo echo;
	public PlayGif pg;
	public Socket roomSocket;
	public User user;
	public Room room;
	// 标识
	public bool flag,rflag,settleflag;
	// 控制移动和操作
	public Vector3 moveForword;	
    float moveSpeed;
	public int act;
	// 打印信息
	public Text ScoreText,TimeText,TopText, TopScoreText, MyTopText, MyScoreText, xtext, ytext,speedtext;
	// 辅助打印信息
	public Text l1x, l1y, l2x, l2y, l3y, l3x, l4x, l4y;
	// 结算信息
	public Text settlementText;

	private float  width, height;  // 边界宽度和高度
	
	// 相机控制
	public Coordinate Camera_pos, Camera_scale;
	public UInt32 cid;
	public float cameSizeMin,cameSizeMax,cameSize;

	float score;

	public GameObject trianglefood, rectfood, hexagonfood, stabfood,sporefood; //可操作的地图对象
	public GameObject back;         // 地图
	public GameObject origenBall,player;    // 圆球
	public GameObject directButton,breakButton,sporeButton;		// 控制操作按钮

	public DateTime preTime ,shuaxinpre,opTime;    // 时间
	public long StartTime;
	// 超参数
	private float chazhi=0.13f;
	// 各个贴图的大小
	Dictionary<string, float> staticData;

	// 对象的映射信息 三角，方形，六边形，刺球
	Dictionary<UInt32, GameObject> mapFood,mapPlay,mapSpore;
	Dictionary<UInt32, Coordinate> SporeBall, PlayBall;
	Dictionary<UInt32, UInt32> mapType;

	List<UInt32> CurrentId;        // 当前存在的对象ID	

	byte[] readBuff;    // 缓冲

	void Awake()
    {
        Application.targetFrameRate = 50;

	}
    // Use this for initialization
    void Start () {
		readBuff = new byte[50000];
		staticData = new Dictionary<string, float>();
		staticData["food"] = 0.13f; //0.065f
		staticData["spore"] = 0.155f;   //0.155f
		staticData["stab"] = 0.8f;  // 160像素=1.6尺寸取一半0.8
		staticData["play"] = 3f;  //  600像素6f取一半3
		staticData["origen"] = 0.435f;  // 87像素,43.5f
		staticData["camera"] = 1.73f;     //play/50 112,14 142.2,10
		InitParameter();
		initelem();
	}
	
	void MyStart()
    {
		user.Name = echo.user.Name;
		user.Id = echo.user.Id;
		user.SkinId = echo.ui.skinId;
		user.Color = new logicProto.Color {
			R = echo.ui.rgba[0], G = echo.ui.rgba[1],
			B= echo.ui.rgba[2],	A= echo.ui.rgba[3],
        };
		room = echo.room;
		flag = echo.flag;
		if (roomSocket==null)
        {
			roomSocket = new Socket(AddressFamily.InterNetwork,
				SocketType.Stream, ProtocolType.Tcp);
			//Connect
			var ip = room.Addr.Split(':')[0];
			int port = int.Parse(room.Addr.Split(':')[1]);
			l1x.text = ip;
			l1y.text = Convert.ToString(port);
			roomSocket.Connect(ip, port);
        }				

		// 请求后端初始化游戏房间
		room.Width = width;
		room.Height = height;
		var request = new Operator
		{
			Operator_ = "init",
			Token = echo.user.Token,
			User = user,
			Room = room
		};

		byte[] bytes;
		using (MemoryStream stream = new MemoryStream())
		{
			// Save the person to a stream
			request.WriteTo(stream);
			bytes = stream.ToArray();
		}
		roomSocket.Send(bytes);
		//Recv
		int count = roomSocket.Receive(readBuff);
		xtext.text = Convert.ToString(count);
		var res = InitResponse.Parser.ParseFrom(readBuff, 0, count);
		DealInit(res);
		rflag = true;
		settleflag = false;
	}

	void DealInit(InitResponse res) 
	{
		if (res.Status != 0)
        {
			xtext.text = "init error";
			return;
        }
		// 设置地图大小符合服务端要求
		InitMap(res.Size);
		// 更新玩家坐标和大小
		var x= res.Player.X;
		var y= res.Player.Y;	
		score = res.Player.Score;
		cid = res.Player.Id;
		var cam_scale = Mathf.Sqrt(score)*staticData["camera"];
		Camera.main.transform.position = new Vector3(x, y, -40);
		Camera.main.orthographicSize = cam_scale;
		Camera_pos = new Coordinate
		{
			X = x,
			Y = y,
			Z = -40,
		};
		Camera_scale = new Coordinate
		{
			X = cam_scale,
			Y = cam_scale,
			Z = 1,
		};
		cameSizeMin *= cam_scale;
		cameSizeMax *= cam_scale;
		GetPlayer(res.Player.Id,res.Player);
		// 更新全局信息
		StartTime = res.Player.CreateTime/1000;
		var bodys = res.Quark;
		Ball it;
		for (int i = 0; i < bodys.Count; ++i)
        {
			it = bodys[i];
			UInt32 id = it.Id;
			var type = it.Type;
            if (type == 1) //food
            {
				GetTri(id, it.X, it.Y);
			}
			else if (type == 2)	
            {
				GetRec(id, it.X, it.Y);
			}
			else if(type == 3)
            {
				GetHex(id, it.X, it.Y);
			}
			else if (type == 4)//刺球
			{
				GetStab(id, it);
			}				
        }
	}

	// Update is called once per frame
	void Update()
    {
		if (Input.GetKey(KeyCode.Escape)) // 退出游戏
		{
			echo.End();
			echo.Login();
		}
		double lu;
		if (echo.playerFlag)
		{
			if (!rflag)
			{
				MyStart();
			}		
		}
		else
		{
			return;
		}
		// 处理输入
		inoutOp();
		lu = (DateTime.Now - preTime).TotalMilliseconds;
		if (lu >= 100)
        {
			preTime = DateTime.Now;
			Update50();			
		}
		DealCamera();
		//Update200();
		// 移动的差值化处理
		// 孢子
		foreach (KeyValuePair<UInt32, GameObject> kvp in mapSpore)
		{
			if (kvp.Value.activeSelf)
			{ UpdatePos(SporeBall[kvp.Key], kvp.Value); }
		}

		// 玩家分身
		foreach (KeyValuePair<UInt32, GameObject> kvp in mapPlay)
		{
			if (kvp.Value.activeSelf) { UpdatePos(PlayBall[kvp.Key], kvp.Value); }
		}
	}
	// 处理输入
	void inoutOp()
    {
		double lu;
		//操作按钮显示更新
		lu = (DateTime.Now - opTime).TotalMilliseconds;
		if (lu >= 1500)
		{
			opTime = DateTime.Now;
			setop();
		}
		else if (directButton.GetComponent<Image>().color.a>0.7&&lu >= 800)
		{
			directButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 0.5f);
		}
		
		// 获取运动方向
		if (moveSpeed >= 4.5f || moveSpeed < 1f) moveSpeed = 3f;
		if (Input.GetMouseButton(0) || Input.GetMouseButton(1))
		{
			MouseFollow();
		}
		if (Input.GetKey(KeyCode.Q))
		{
			act = 1;
		}
		if (Input.GetKey(KeyCode.Space)|| Input.GetKey(KeyCode.E))
		{
			act = 2;
		}	
		if (Input.GetKey(KeyCode.UpArrow) || Input.GetKey(KeyCode.W))
		{
			moveForword.y += 1;
		}
		if (Input.GetKey(KeyCode.DownArrow) || Input.GetKey(KeyCode.S))
		{
			moveForword.y += -1;
		}
		if (Input.GetKey(KeyCode.LeftArrow) || Input.GetKey(KeyCode.A))
		{
			moveForword.x += -1;
		}
		if (Input.GetKey(KeyCode.RightArrow) || Input.GetKey(KeyCode.D))
		{
			moveForword.x += 1;
		}
		moveSpeed = Mathf.Clamp(moveSpeed, 1, 5);
		moveForword.x = Mathf.Clamp(moveForword.x, -3, 3);
		moveForword.y = Mathf.Clamp(moveForword.y, -3, 3);
		if (moveForword.x == 0)
		{
			moveForword.x = 0.1f;
		}
		if (moveForword.y == 0)
		{
			moveForword.y = 0.1f;
		}
		// 方向归一化
		var mlen = Mathf.Sqrt(moveForword.x * moveForword.x + moveForword.y * moveForword.y);
		moveForword.x = moveForword.x / mlen;
		moveForword.y = moveForword.y / mlen;
		ytext.text = "y: " + Convert.ToString(moveForword.y);
		xtext.text = "x: " + Convert.ToString(moveForword.x);
		speedtext.text = "s: " + Convert.ToString(moveSpeed);
	}
	// 每100ms刷新一次
	void Update50 () {	
		var request = new Operator
		{			
			Operator_ = "update",
			//Token = echo.user.Token,
			User = user,
			Room = room,
			Act = new Act
			{
				MoveX = moveForword.x,
				MoveY = moveForword.y,
				MoveS = moveSpeed,
				Type = act,
			},
		};
		act = 0;
		byte[] bytes;
		using (MemoryStream stream = new MemoryStream())
		{
			request.WriteTo(stream);
			bytes = stream.ToArray();
		}
		roomSocket.Send(bytes);
		//Recv		
		int count = roomSocket.Receive(readBuff);
		//xtext.text ="count: "+Convert.ToString(count);
		while (count == 50000)
        {
			count = roomSocket.Receive(readBuff);
			if (count != 50000)
            {
				return;
            }
		}
		var res = UpdateResponse.Parser.ParseFrom(readBuff, 0, count);
		DealUpdate(res);
	}

	void DealUpdate(UpdateResponse res)
	{
		if (res.Status > 0)
		{
			xtext.text = res.Msg;
			return;
		}
		if (res.Status == -1)
		{
			Settlement(res);
			return;
		}

		// 修正后的摄像机		
		score = res.CamPos.Score;
		cid = res.CamPos.Id;
		var posCa = new Vector3(res.CamPos.X, res.CamPos.Y, -10);		
		posCa.x = Mathf.Clamp(posCa.x, 0, 200);
		posCa.y = Mathf.Clamp(posCa.y, 0, 200);
		
		Camera_pos = new Coordinate
		{
			X = posCa.x,
			Y = posCa.y,
			Z = 0,
		};
		Camera_scale = new Coordinate
		{
			X = res.CamPos.Sx,
			Y = res.CamPos.Sy,
			Z = 0,
		};

		// 返回的foods、spore、player
		CurrentId.Clear();
		// 粮食 foods (food, stab)
		for (int i = 0; i < res.Foods.Count; ++i)
		{
			var food = res.Foods[i];
			var id = food.Id;
			CurrentId.Add(id);
			var ty = food.Type;
			if (!mapFood.ContainsKey(id))
			{
				switch (ty)
				{
					case 1: GetTri(food.Id, food.X, food.Y); ; break;
					case 2: GetRec(food.Id, food.X, food.Y); ; break;
					case 3: GetHex(food.Id, food.X, food.Y); ; break;
					case 4: GetStab(food.Id, food); break;
				}
			}
		}
		// 孢子 spore
		for (int i = 0; i < res.Spore.Count; ++i)
		{
			var spore = res.Spore[i];
			var id = spore.Id;
			CurrentId.Add(id);
			if (!mapSpore.ContainsKey(id))
			{
				GetSpore(id, spore);
			}
		}
		// 玩家 player
		for (int i = 0; i < res.Players.Count; ++i)
		{
			var player = res.Players[i];
			var id = player.Id;
			CurrentId.Add(id);
			if (mapPlay.ContainsKey(id))
			{
				float sc = Score2R(player.Score) / staticData["play"];
				
				var tmpgm = mapPlay[id].transform.Find("name").gameObject;
				if (player.Score <= 40)
				{
					tmpgm.SetActive(false);
                }
                else
                {
					tmpgm.SetActive(true);
				}
				mapPlay[id].transform.localScale = new Vector3(sc, sc, 1);
				if (!mapPlay[id].activeSelf) mapPlay[id].SetActive(true);
				PlayBall[id] = new Coordinate
				{
					X = player.Nx,
					Y = player.Ny,
					Z = Score2C(player.Score),
				};
			}
			else
			{
				GetPlayer(id, player);
			}
		}

		// 待删除的对象
		CurrentId.Sort();
		DealGame();

		// 更新排行榜
		var lu =(DateTime.Now-shuaxinpre).TotalMilliseconds; 
		if (lu >900)
        {
			string s = "    " + res.Top.Pair[0].Name + "\n";
			string scs = (int)res.Top.Pair[0].Val + " \n";
			l4x.text = "排行榜人数";
			l4y.text = Convert.ToString(res.Top.Pair.Count);
			for (int i = 2; i <= res.Top.Pair.Count; i++)
			{
				var val = res.Top.Pair[i - 1];
				scs = scs + Convert.ToString((int)val.Val) + " \n";
				if (i < 10)
					s = s + " ";
				s = s + Convert.ToString(i) + "." + val.Name + "\n";
			}
			TopText.text = s;
			TopScoreText.text = scs;
			MyTopText.text = res.Top.MyTop + "." + user.Name;
			MyScoreText.text = Convert.ToString((int)score) + " ";
			shuaxinpre = DateTime.Now;
		}
		PostUpdate(res.CamPos);
	}

	void GetTri(UInt32 id,float x,float y)
    {
		var food = Instantiate(trianglefood);
		food.transform.position = new Vector3(x, y, -0.5f);
		food.transform.localScale = new Vector3(2, 2,1);
		food.SetActive(true);
		mapFood[id] = food;
		mapType[id] = 1;
	}
	void GetRec(UInt32 id, float x, float y)
	{
		var food = Instantiate(rectfood);
		food.transform.position = new Vector3(x, y, -0.5f);
		food.transform.localScale = new Vector3(2, 2, 1);
		food.SetActive(true);
		mapFood[id] = food;
		mapType[id] = 2;
	}

	void GetHex(UInt32 id, float x, float y)
	{
		var food = Instantiate(hexagonfood);
		food.transform.position = new Vector3(x, y, -0.5f);
		food.transform.localScale = new Vector3(2, 2, 1);
		food.SetActive(true);
		mapFood[id] = food;
		mapType[id] = 3;
	}

	void GetStab(UInt32 id,Ball b)
	{
		var food = Instantiate(stabfood);
		food.transform.position = new Vector3(b.X, b.Y, Score2C(b.Score));
		var sc = Score2R(b.Score)/staticData["stab"];
		food.transform.localScale = new Vector3(sc, sc, 1);
		food.SetActive(true);
		mapFood[id] = food;
		mapType[id] = 4;
	}

	void GetPlayer(UInt32 id,Ball b)
	{
		var Skin = b.SkinId;
		GameObject pl = Instantiate(player);
		var skinSprite=pl.transform.Find("skinSprite").GetComponent<SpriteRenderer>();
		var nameMesh=pl.transform.Find("name").GetComponent<TextMesh>();
		skinSprite.sprite = echo.ui.imgMap[Skin];
		nameMesh.text = b.Name;
		if (b.Score <= 40)
        {
			pl.transform.Find("name").gameObject.SetActive(false);
		}
				
		if (b.Color == null)
		{
			b.Color = user.Color;
		}
		var color = new UnityEngine.Color((float)b.Color.R / 255, (float)b.Color.G, (float)b.Color.B, (float)b.Color.A);
		nameMesh.color = color;
		if(Skin==0)skinSprite.color=color;		

		float sX=staticData["play"]*200/echo.ui.skinMap[Skin].Width;
		float sY= staticData["play"] * 200/echo.ui.skinMap[Skin].Height;
		float sc=Score2R(b.Score) / staticData["play"];

		pl.transform.position = new Vector3(b.X, b.Y, Score2C(b.Score));
		skinSprite.transform.localScale = new Vector3(sX, sY, 1);
		pl.transform.localScale = new Vector3(sc, sc, 1);
		pl.SetActive(true);
		mapPlay[id] = pl;
		PlayBall[id] = new Coordinate
		{
			X=b.Nx,
			Y=b.Ny,
			Z=0f,
		};
		mapType[id] = 0;
	}

	void GetSpore(UInt32 id, Ball b)
	{
		var Skin = b.SkinId;
		GameObject spore;
		float sc;
		if (Skin == 0)
		{
			spore = Instantiate(origenBall);
			if (b.Color == null){
				b.Color = user.Color;
            }
			var color = new UnityEngine.Color((float)b.Color.R / 255, (float)b.Color.G, (float)b.Color.B, (float)b.Color.A);
			//spore.GetComponent<SpriteRenderer>().material.SetColor("_Color", new UnityEngine.Color(b.Color.R, b.Color.G, b.Color.B, b.Color.A));
			spore.GetComponent<SpriteRenderer>().color = color;
			sc = Score2R(b.Score) / staticData["origen"];
		}
		else
		{
			spore = Instantiate(sporefood);
			sc = Score2R(b.Score) / staticData["spore"];
		}
		spore.transform.position = new Vector3(b.X, b.Y, Score2C(b.Score));
		spore.transform.localScale = new Vector3(sc, sc, 1);
		spore.SetActive(true);
		mapSpore[id] = spore;
		SporeBall[id] = new Coordinate{ X=b.Nx, Y=b.Ny,Z= Score2C(b.Score) };
		mapType[id] = 5;
	}

	void PostUpdate(CameraPos cam)
    {
		ScoreText.text = "得分：" + Convert.ToString((int)score);
		var nowTime = cam.Curtime- StartTime;
		// 倒计时 
		if (nowTime > 180||nowTime<0)
        {
			StartTime = cam.Curtime;
        }
		nowTime = 180 - nowTime;
		l2x.text = Convert.ToString(StartTime);
		l2y.text = Convert.ToString(cam.Curtime);

		TimeText.text = Convert.ToString(nowTime / 60) + ":"+ Convert.ToString(nowTime % 60);
		PrintText();
	}

	void Settlement(UpdateResponse ret)
    {
		settleflag = true;		
		rflag = false;
		echo.room2settle();
		settlementText.text = "  游戏结束！\n" +"  "+Convert.ToString(ret.Msg)+"\n"+"  当前等级："+ Convert.ToString(ret.User.Level) + "\n";
		echo.ulevel.text=echo.userLevel.text = "等级:" + Convert.ToString(ret.User.Level);
		echo.userExp.text = "Exp:" + Convert.ToString(ret.User.Experience);
		CurrentId.Clear();
		DealGame();	
	}

	void UpdatePos(Coordinate ball, GameObject ob)
    {		
		var modVal= new Vector3(Mathf.Lerp(ob.transform.position.x, ball.X, chazhi), Mathf.Lerp(ob.transform.position.y, ball.Y, chazhi),ball.Z);
		modVal.x = Mathf.Clamp(modVal.x, 0, 200);
		modVal.y = Mathf.Clamp(modVal.y, 0, 200);
		ob.transform.position = modVal;		
    }

	void DealCamera()
    {
		if (settleflag)
        {
			return;
        }
		Vector3 modVal;	

		float x = Camera.main.transform.position.x, y = Camera.main.transform.position.y,size =Camera.main.orthographicSize;		
		if (Camera_pos.X - x>32|| Camera_pos.X - x<-32|| Camera_pos.Y - y > 18 || Camera_pos.Y - y < -18)
        {
			modVal = new Vector3(Camera_pos.X,Camera_pos.Y,-40);
        }
        else
        {
			modVal = new Vector3(Mathf.Lerp(x, Camera_pos.X, chazhi * 0.11f), Mathf.Lerp(y, Camera_pos.Y, chazhi * 0.06f), -40);
		}
		modVal.x = Mathf.Clamp(modVal.x, 0,200);
		modVal.y = Mathf.Clamp(modVal.y, 0,200);
		Camera.main.transform.position = modVal;
		//Camera.main.transform.localScale = new Vector3(Camera_scale.X, Camera_scale.Y, 1);
		var k =Math.Max(Camera_scale.X*9/16, Camera_scale.Y);
		k = size*Mathf.Clamp(k/ size,0.995f,1.005f);
		Camera.main.orthographicSize = k;
	}

	void PrintText()
    {
		// 当前坐标
		l1x.text = "镜头x "+ Convert.ToString(Camera_pos.X);
		l1y.text = Convert.ToString(Camera_pos.Y);
		if (CurrentId.Count > 0)
        {
			l3x.text = "key: " + Convert.ToString(CurrentId[0]);
			l3y.text = Convert.ToString(CurrentId[CurrentId.Count-1]);
        }		
	}
	void DealGame()
    {
		var cnt = 0;
		List<UInt32> tempDel = new List<UInt32>();
		foreach (KeyValuePair<UInt32, GameObject> kvp in mapFood)
		{
			var id = kvp.Key;
			if (!findList(id))
			{
				tempDel.Add(id);
			}
		}
		cnt += tempDel.Count;
		foreach (UInt32 id in tempDel)
		{
			Destroy(mapFood[id]);
			mapFood.Remove(id);
		}
		tempDel.Clear();
		foreach (KeyValuePair<UInt32, GameObject> kvp in mapSpore)
		{
			var id = kvp.Key;
			if (!findList(id))
			{
				tempDel.Add(id);
				SporeBall.Remove(id);
			}
		}
		cnt += tempDel.Count;
		foreach (UInt32 id in tempDel) { Destroy(mapSpore[id]); mapSpore.Remove(id); }
		tempDel.Clear();
		foreach (KeyValuePair<UInt32, GameObject> kvp in mapPlay)
		{
			var id = kvp.Key;
			if (!findList(id))
			{
				tempDel.Add(id);
				PlayBall.Remove(id);
			}
		}
		cnt += tempDel.Count;
		foreach (UInt32 id in tempDel) { Destroy(mapPlay[id]); mapPlay.Remove(id); }
		l3x.text = "删除 " + Convert.ToString(cnt);		
	}

	//操作控制
	public void sporeB()
    {
		act = 1;
		sporeButton.GetComponent<Image>().color = new UnityEngine.Color(1,1, 1,1);
	}
	public void breakB()
	{
		act = 2;
		breakButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 1);
	}

	public void playskin()
    {
		pg.playskin(mapPlay[cid]);
	}
	/// <summary>
	/// 获取鼠标点击坐标的方法
	/// </summary>
	public void MouseFollow()
	{
		//获取游戏对象在世界坐标中的位置，并转换为屏幕坐标；
		var screenPosition = Camera.main.WorldToScreenPoint(directButton.transform.position);
		var wi = Screen.width;
		var he=Screen.height;

		//获取鼠标在场景中坐标
		var mousePositionOnScreen = Input.mousePosition;		
		//让鼠标坐标的Z轴坐标 等于 场景中游戏对象的Z轴坐标
		mousePositionOnScreen.z = screenPosition.z;
		//将鼠标的屏幕坐标转化为世界坐标
		var mousePositionInWorld = Camera.main.ScreenToWorldPoint(mousePositionOnScreen);

		// 是否播放gif
		if (mapPlay[cid].activeSelf&&distv3(mapPlay[cid].transform.position, mousePositionInWorld) <2.5)
        {
			Debug.Log("播放");
			pg.playskin(mapPlay[cid]);
		}
		if(mousePositionOnScreen.x> wi * 2 / 3 && (mousePositionOnScreen.y>he*3/4|| mousePositionOnScreen.y < he / 4))
        {
			return;
        }		
		var dir = mousePositionInWorld- Camera.main.transform.position;
		dir.z = 0;

		//将游戏对象的坐标改为鼠标的世界坐标，物体跟随鼠标移动
		directButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 1);
		directButton.transform.position = mousePositionInWorld;
		
		//移动方向
		moveForword = (moveForword + dir) / 2;
	}

	void setop()
    {
		sporeButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 0.5f);
		breakButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 0.5f);
		directButton.GetComponent<Image>().color = new UnityEngine.Color(1, 1, 1, 0);
	}

	void InitMap(int size)
    {
		var rectTransform = back.GetComponent<RectTransform>();
		var x = size / rectTransform.rect.xMax;
		var y = size / rectTransform.rect.yMax;
		back.transform.localScale = new Vector3(x,y,1);
		back.transform.position = new Vector3(size,size,0);
		width = rectTransform.rect.xMax * rectTransform.localScale.x;    // 计算可移动范围
		height = rectTransform.rect.yMax * rectTransform.localScale.y;

		l4x.text = "背景: " + Convert.ToString(width) + "," + Convert.ToString(rectTransform.localScale.x);
		l4y.text = Convert.ToString(height) + "," + Convert.ToString(rectTransform.localScale.y);

		rectTransform = Camera.main.GetComponent<RectTransform>();
		width = rectTransform.rect.xMax * rectTransform.localScale.x;    // 计算可移动范围
		height = rectTransform.rect.yMax * rectTransform.localScale.y;
		l2x.text = "相机: " + Convert.ToString(width) + "," + Convert.ToString(rectTransform.localScale.x);
		l2y.text = Convert.ToString(Camera.main.orthographicSize);
	}

	void InitParameter()
    {
		moveForword = new Vector3(0.5f, 0.5f, 0);
		moveSpeed = 1;
		cameSizeMax = 10.0f;
		cameSizeMin = 1f;

		mapFood = new Dictionary<UInt32, GameObject>();
		mapSpore = new Dictionary<UInt32, GameObject>();
		mapPlay = new Dictionary<UInt32, GameObject>();
		mapType = new Dictionary<UInt32, UInt32>();
		SporeBall = new Dictionary<UInt32, Coordinate>();
		PlayBall = new Dictionary<UInt32, Coordinate>();
		
		CurrentId = new List<UInt32>();

		user = new User {SkinId=1, };

		settleflag=rflag = false;
		opTime=shuaxinpre = preTime = DateTime.Now;
	}

	void initelem()
    {
		trianglefood.SetActive(false); rectfood.SetActive(false); hexagonfood.SetActive(false); 
		stabfood.SetActive(false); sporefood.SetActive(false);	origenBall.SetActive(false); player.SetActive(false);
	}
	bool findList(UInt32 key) // 二分查找
    {
		if (CurrentId.Count==0|| CurrentId[0] > key)
			return false;
		var l = 0;
		var r = CurrentId.Count - 1;
        while (l <= r)
        {
			var mid = (l + r) / 2;
			if (CurrentId[mid] > key)
            {
				r = mid -1;
            }else if (CurrentId[mid] < key)
            {
				l = mid+1;
            }
            else
            {
				return true;
			}
        }
		
		return false;
    }

	// 获取游戏对象尺寸
	void getGameObjectSize(GameObject gb)
    {
		var rectTransform = gb.GetComponent<RectTransform>();
		var x = rectTransform.rect.width;
		var y = rectTransform.rect.height;
		var sx = rectTransform.localScale.x;
		var sy = rectTransform.localScale.x;
	}

	// 分数计算半径
	float Score2R(float score)
    {
		if (score > 30000)
        {
			score = 30000;
        }
		return Mathf.Sqrt(0.042f * score + 0.15f);
    }
	// 分数计算图层
	float Score2C(float score)
	{
		var ra = Score2R(score);
		if (ra > 35.5)
			return -35.5f;

		return -ra;
	}

	// 计算两个vector3的距离
	float distv3(Vector3 a,Vector3 b)
    {
		return Mathf.Sqrt((a.x - b.x) * (a.x - b.x) + (a.y - b.y) * (a.y - b.y));
	}
}
