using System.Collections;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UI;
using System;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using System.Security.Cryptography;

public class UIcontroller : MonoBehaviour {	

	public GameObject panellogin, panelmenu,panelskin, panelselect, panelmatch, panelroom, panelsettle;   // 图层
	public GameObject imgLogin, imgMenu,imgSkin, imgSelect, imgMatch, imgRoom, imgSettle;           // 背景板
	public GameObject skinshow,myskinshow,todoshow,imgpika,imgkena,imgworker,imgmoon;		// 皮肤展示,敬请期待
	public GameObject register1input, loginB, register2input, repasswdinput;        // 登录按钮
	public GameObject operatorB,opshowB;			// 说明书
	public GameObject debugB,testText;          // 是否打印测试条
	public GameObject topButten,topPanel;				// 是否展示排行榜

	public InputField inputR, inputG, inputB, inputA,inputPath;		//颜色输入框
	public Text skinName,debugT,toptitle,todotext,titletext;							// 皮肤名称,debug显示,排行榜

	private const float REFERENCE_RESOLUTION_WIDTH = 1920f; // 参考分辨率宽度
	private const float REFERENCE_RESOLUTION_HEIGHT = 1080f; // 参考分辨率高度

	public UInt16[] rgba;			// RGBA 255
	public float[] RGBA;			//RGBA float0-1
	public uint skinId;				// 皮肤ID
	public string[] nameList;       // 随机名称库
	public LocalDialog localDia;	// 文件选择框
	// 读取的图片提高贴图
	public Dictionary<uint, System.Drawing.Image> skinMap;		
	public Dictionary<uint, Sprite> imgMap;
	public Dictionary<uint, Texture2D> t2DMap;
	
	void Start () {
		initback();
		initskin();
		initname();
		localDia = new LocalDialog();
	}

	// 获取皮肤直接从文件读取，而不是都放到项目里
	public void path2skin()
	{
		localDia.OpenDirectory("png");
		var s = localDia.filePath.Split('.');
		inputPath.text = localDia.filePath;
		//var s = inputPath.text.Split('.');
		if (s.Length < 2)
		{
			todotext.text = "地址错误！";
			return;
		}
		var suffit = s[s.Length - 1];
		if (suffit != "png")
		{
			todotext.text = "只支持png图片！";
			return;
		}
		skinMap[5] = System.Drawing.Image.FromFile(inputPath.text);
		var tex = img2text2D(skinMap[5]);
		t2DMap[5] = tex;
		imgMap[5] = Sprite.Create(tex, new Rect(0.0f, 0.0f, tex.width, tex.height), new Vector2(0.5f, 0.5f));
		changeImg(5);
		skinshow.SetActive(true);
	}
	// 生成随机颜色的皮肤
	public void randomSkin()
	{
		skinId = 0;
		randomcolor();
		myskinshow.GetComponent<UnityEngine.UI.Image>().sprite = imgMap[0];
		skinshow.GetComponent<UnityEngine.UI.Image>().sprite = imgMap[0];
		inputR.text = Convert.ToString(rgba[0]);
		inputG.text = Convert.ToString(rgba[1]);
		inputB.text = Convert.ToString(rgba[2]);
		inputA.text = Convert.ToString(rgba[3]);
		showMySkin();
	}

	// 生成随机颜色
	void randomcolor()
    {
		for (int i = 0; i < 4; i++)
		{
			RNGCryptoServiceProvider csp = new RNGCryptoServiceProvider();
			byte[] byteCsp = new byte[3];
			csp.GetBytes(byteCsp);
			rgba[i] = (ushort)(BitConverter.ToUInt16(byteCsp, 0) % 255);
            if (i == 3)
            {
				rgba[i] = (ushort)(BitConverter.ToUInt16(byteCsp, 0) % 125);
				rgba[i] += 130;
            }				
		}
	}
	// 从input获取颜色输入
	public void inputskin()
	{
		rgba[0] = Convert.ToUInt16(inputR.text);
		rgba[1] = Convert.ToUInt16(inputG.text);
		rgba[2] = Convert.ToUInt16(inputB.text);
		rgba[3] = Convert.ToUInt16(inputA.text);
		skinId = 0;
		showMySkin();
	}
	void initskin()
    {
		rgba = new UInt16[4];
		RGBA = new float[4];
		skinMap = new Dictionary<uint, System.Drawing.Image>();
		t2DMap = new Dictionary<uint, Texture2D>();
		imgMap = new Dictionary<uint, Sprite>();
		skinMap[0]=System.Drawing.Image.FromFile(Application.dataPath +"/source/front/转盘.png");
		skinMap[1]=System.Drawing.Image.FromFile(Application.dataPath + "/source/front/kenan.png");
		skinMap[2]=System.Drawing.Image.FromFile(Application.dataPath + "/source/front/dasda.png");
		skinMap[3]=System.Drawing.Image.FromFile(Application.dataPath + "/source/front/nvVTSFuKGa.png");
		skinMap[4]=System.Drawing.Image.FromFile(Application.dataPath + "/source/front/duitang.png");
		foreach(KeyValuePair<uint, System.Drawing.Image> kv in skinMap)
        {
			var tex= img2text2D(kv.Value);
			t2DMap[kv.Key] = tex;
			imgMap[kv.Key] =Sprite.Create(tex, new Rect(0.0f, 0.0f, tex.width, tex.height), new Vector2(0.5f, 0.5f));
		}
		randomcolor();
	}

	private Texture2D img2text2D(System.Drawing.Image image)
    {
		Bitmap framBitmap = new Bitmap(image.Width, image.Height);
		using (System.Drawing.Graphics graphic = System.Drawing.Graphics.FromImage(framBitmap))
		{
			graphic.DrawImage(image, Point.Empty);
		}
		Texture2D frameTexture2D = new Texture2D(framBitmap.Width, framBitmap.Height, TextureFormat.ARGB32, true);
		frameTexture2D.LoadImage(Bitmap2Byte(framBitmap));
		return frameTexture2D;
	}

	private byte[] Bitmap2Byte(Bitmap bitmap)
	{
		using (MemoryStream stream = new MemoryStream())
		{
			// 将bitmap 以png格式保存到流中
			bitmap.Save(stream, ImageFormat.Png);
			// 创建一个字节数组，长度为流的长度
			byte[] data = new byte[stream.Length];
			// 重置指针
			stream.Seek(0, SeekOrigin.Begin);
			// 从流读取字节块存入data中
			stream.Read(data, 0, Convert.ToInt32(stream.Length));
			return data;
		}
	}

	public void skinkena()
    {
		skinName.text = "柯南";
		changeImg(1);
		backkena();
	}
	public void skinpika()
	{
		skinName.text = "皮卡丘";
		changeImg(2);
		backpika();
	}
	public void skinworker()
	{
		skinName.text = "打工人";
		changeImg(3);
		backworker();
	}
	public void skinmoon()
	{
		skinName.text = "月夜";
		changeImg(4);
		backmoon();
	}

	private void Awake()
	{
		AdjustScreen();
	}

	private void AdjustScreen()
	{
		float screenWidth = Screen.width; // 获取当前屏幕宽度
		float screenHeight = Screen.height; // 获取当前屏幕高度
		float screenAspectRatio = screenWidth / screenHeight; // 获取当前屏幕宽高比

		float referenceAspectRatio = REFERENCE_RESOLUTION_WIDTH / REFERENCE_RESOLUTION_HEIGHT; // 获取参考宽高比

		// 如果当前宽高比小于参考宽高比，说明是窄屏，需要将摄像机视野进行调整
		if (screenAspectRatio < referenceAspectRatio)
		{
			float targetWidth = screenHeight / REFERENCE_RESOLUTION_HEIGHT * REFERENCE_RESOLUTION_WIDTH;
			float widthDiff = screenWidth - targetWidth;

			Camera.main.orthographicSize *= targetWidth / screenWidth;

			Camera.main.transform.position += new Vector3(widthDiff / 2f, 0f, -10f);
		}
		// 如果当前宽高比大于参考宽高比，说明是宽屏，需要调整摄像机的位置
		else if (screenAspectRatio > referenceAspectRatio)
		{
			float targetHeight = screenWidth / REFERENCE_RESOLUTION_WIDTH * REFERENCE_RESOLUTION_HEIGHT;
			float heightDiff = screenHeight - targetHeight;

			Camera.main.transform.position += new Vector3(0f, heightDiff / 2f, -10f);
		}
	}
	public void initback()
	{
		skin(false);
		select(false);
		match(false);
		room(false);
		settle(false);
		untodo();
		opshowB.SetActive(false);
		testText.SetActive(false);
		changeback();
		debugT.text = " ";
	}
	public void login(bool b)
	{
		panellogin.SetActive(b);
		imgLogin.SetActive(b);
		if (b)
        {
			Camera.main.transform.position = 
				new Vector3(imgLogin.transform.position.x, imgLogin.transform.position.y, -10);
		}
	}
	public void menu(bool b)
	{
		panelmenu.SetActive(b);
		imgMenu.SetActive(b);
		if (b)
		{
			Camera.main.transform.position = 
				new Vector3(imgMenu.transform.position.x, imgMenu.transform.position.y, -10);
        }
        else
        {
			untodo();
		}
	}
	public void select(bool b)
	{
		panelselect.SetActive(b);
		imgSelect.SetActive(b);
		if (b)
		{
			Camera.main.transform.position =
				new Vector3(imgSelect.transform.position.x, imgSelect.transform.position.y, -10);
		}
	}
	public void match(bool b)
	{
		panelmatch.SetActive(b);
		imgMatch.SetActive(b);
		if (b)
		{
			Camera.main.transform.position =
				new Vector3(imgMatch.transform.position.x, imgMatch.transform.position.y, -10);
		}
	}
	public void room(bool b)
	{
		panelroom.SetActive(b);
		imgRoom.SetActive(b);
	}
	public void settle(bool b)
	{
		panelsettle.SetActive(b);
		imgSettle.SetActive(b);
		if (b)
		{
			Camera.main.orthographicSize = 10;
			Camera.main.transform.position =new Vector3(imgSettle.transform.position.x, imgSettle.transform.position.y, -10);
		}		
	}
	// 控制对象旋转
	public void xuanzhuan(GameObject gb)
    {		
		Vector3 angles = gb.transform.localEulerAngles;
		angles.z += 360 * Time.deltaTime;
		gb.transform.localEulerAngles = angles;
    }
	// 说明书
	public void opshow()
    {
		opshowB.SetActive(!opshowB.activeSelf);
    }

	public void debugText()
    {
		testText.SetActive(!testText.activeSelf);
		if (testText.activeSelf) {
			debugT.text = "✔";
        }
        else
        {
			debugT.text = "✘";			
		}		
	}
	public void skin(bool b)
    {
		panelskin.SetActive(b);
		imgSkin.SetActive(b);
		if (b)
		{
			Camera.main.transform.position =
				new Vector3( imgSkin.transform.position.x, imgSkin.transform.position.y, -10);
		}
	}

	public void pressbtnA()
    {
		register1input.SetActive(false);
		loginB.SetActive(false);
		register2input.SetActive(true);
		repasswdinput.SetActive(true);
	}

	public void rePressbtnA()
	{
		register1input.SetActive(true);
		loginB.SetActive(true);
		register2input.SetActive(false);
		repasswdinput.SetActive(false);
	}

	void showMySkin()
    {
		for (int i = 0; i < 4; i++)
		{
			RGBA[i] = (float)rgba[i] / 255;
		}
		myskinshow.GetComponent<UnityEngine.UI.Image>().color = new UnityEngine.Color(
			RGBA[0], RGBA[1], RGBA[2], RGBA[3]);
		skinshow.GetComponent<UnityEngine.UI.Image>().color = new UnityEngine.Color(
			RGBA[0], RGBA[1], RGBA[2], RGBA[3]);
	}

	public void todo()
    {
		todoshow.SetActive(true);
	}

	public void untodo()
	{
		todoshow.SetActive(false);
		titletext.gameObject.SetActive(false);
		todotext.text = "敬 请 期 待...";
	}

	public void topB()
    {
		topPanel.SetActive(!topPanel.activeSelf);
		if (topPanel.activeSelf)
        {
			toptitle.text = "排行榜﹀";
        }
        else
        {
			toptitle.text = "排行榜︿";
		}
	}
	void changeImg(uint id)
    {
		skinId = id;
		myskinshow.GetComponent<UnityEngine.UI.Image>().sprite = imgMap[id];
		myskinshow.GetComponent<UnityEngine.UI.Image>().color = new UnityEngine.Color(1f, 1f, 1f, 1f);
		skinshow.GetComponent<UnityEngine.UI.Image>().sprite = imgMap[id];
		skinshow.GetComponent<UnityEngine.UI.Image>().color = new UnityEngine.Color(1f, 1f, 1f, 1f);
	}

	void backpika()
    {
		changeback();
		imgpika.SetActive(true);
	}
	void backkena()
	{
		changeback();
		imgkena.SetActive(true);
	}
	void backworker()
	{
		changeback();
		imgworker.SetActive(true);
	}
	void backmoon()
	{
		changeback();
		imgmoon.SetActive(true);
	}
	void changeback()
    {
		imgpika.SetActive(false);
		imgkena.SetActive(false);
		imgworker.SetActive(false);
		imgmoon.SetActive(false);
	}

	void initname()
	{
		nameList = new string[]{"皮皮虾","剑仙","寂寞","守护鲲鲲","尊嘟笨嘟","图森破","图样","狂拽酷炫","逆转乾坤","神级操作",
	  "压力山大","十一","陈皮皮","路三世","熊人领","达令","神秘队友","隐秘杀机","躺赢局局长","暴走战神","疯狂小队",
	  "爱丽丝","大作战","隔壁老王","查水表","几技","秒杀全场","荣耀王者","神仙队友","传奇战神","杀戮机器","无限嚣张",
	  "狂暴之路","全场最佳", "伍六七","网恋奔现","西瓜皮","保熟","疯狂杀戮者","勇者无畏","卓越非凡","荣耀辉煌","火力全开",
	  "峡谷扛把子","稳住别浪", "胜利在望",  "顶尖高手", "独孤求败","神级战士","狂暴战士","无敌战神","一击必杀","神级法师",
	  "神秘战士","狂暴之力","无敌战神","狂野猎手","死亡之影","黑暗势力","全场焦点","先知无敌","神仙下凡","所向披靡",
	  "开挂人生","无限嚣张","战无不胜","疯狂杀戮者"};
	}

}
