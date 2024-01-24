using UnityEngine;
using System.Net.Sockets;
using UnityEngine.UI;
using Google.Protobuf;
using logicProto;
using System.IO;
using System;
using System.Security.Cryptography;

public class Echo : MonoBehaviour {

	// 定义套接字
	public Socket socket;
	public UIcontroller ui;
	// UGUI
	public InputField nameinput, passwdinput, passwdinput1, nickNameinput;
	public Text outText, popup, selectText, selectText2, selectime;
	public bool flag = false, playerFlag = false, matchflag = false, loginflag = false;
	public GameObject popwindows, player; // 弹窗,player

	public Room room;
	// 用户信息 
	public UserInfo user;
	public Text userName, userLevel, userExp, uid, uName, ulevel;

	public DateTime matchTime, preTime, loginTime;
	public bool gifbool;

	// Use this for initialization
	void Start()
	{
		initlogin();
		// Socket
		socket = new Socket(AddressFamily.InterNetwork,
			SocketType.Stream, ProtocolType.Tcp);
		//Connect
		socket.Connect("192.168.96.207", 8080);
		flag = true;

		//Login();		
		selectText.text = "请 选 择 游 戏 模 式";
		gifbool = false;
		loginTime = DateTime.Now;
		Application.runInBackground = true;
	}

	// 后处理
	Response PostDeal(Operator request)
	{
		byte[] bytes;
		using (MemoryStream stream = new MemoryStream())
		{
			// Save the person to a stream
			request.WriteTo(stream);
			bytes = stream.ToArray();
		}
		socket.Send(bytes);
		//Recv
		byte[] readBuff = new byte[1024];
		int count = socket.Receive(readBuff);
		return Response.Parser.ParseFrom(readBuff, 0, count);
	}
	// 登录
	public void Login()
	{
		// Send
		if (!flag) Start();
		string name = nameinput.text, passwd = passwdinput.text;
		//string name = "fht";
		//string passwd = "123456";
        if (!checkedstring(name, passwd))
        {
			return;
        }

		var request = new Operator
		{
			Operator_ = "login",

			User = new User
			{
				Name = name,
				Passwd = passwd
			}

		};
		var response = PostDeal(request);
		if (response.Status == 0)
		{
			outinfo("登录成功!");
			user = response.Userinfo;
			//ui.Manager.Loadui.(1);
			Login2menu();
		}
		else
		{
			outinfo(response.Msg);
		}
	}

	// 注册账号
	public void Account()
	{
		if (!flag) Start();
		string name = nameinput.text, passwd = passwdinput.text, passwd1 = passwdinput1.text;
		if (passwd != passwd1)
        {
			outinfo("两次输入的密码不一样！");
			return;
        }
		if (!checkedstring(name,passwd))
        {
			return;
        }
		var request = new Operator
		{
			Operator_ = "account",

			User = new User
			{
				Name = name,
				Passwd = passwd,
				RePasswd = passwd1,
			}

		};
		var response = PostDeal(request);
		if (response.Status == 0)
		{
			outinfo("注册成功!");
			user = response.Userinfo;
			Login2menu();
		}
		else
		{
			popwindows.SetActive(true);
			popup.text = "注册失败！！请重试";
			outinfo("");
		}
	}

	public void GetAll()
	{
		if (!flag) Start();
		var request = new Operator
		{
			Operator_ = "getAll"
		};
		var response = PostDeal(request);
		if (response.Status == 0)
		{
			var list = response.UserList;
			string s = "";
			for (int i = 0; i < list.Count; ++i) {
				s = s + "," + list[i];
			}
			outinfo(s);
		}
		else
		{
			outinfo(response.Msg);
		}
	}

	// 选择游戏模式-匹配
	public void MetaTest()
	{
		Select2match();
		selectText.text = "匹 配 中...";
		var request = new Operator
		{
			Operator_ = "meta",
			Token = user.Token,
			User = new User
			{
				Name = user.Name,
				Id = user.Id,
			}
		};

		byte[] bytes;
		using (MemoryStream stream = new MemoryStream())
		{
			// Save the person to a stream
			request.WriteTo(stream);
			bytes = stream.ToArray();
		}

		socket.Send(bytes);
		matchflag = true;
		preTime = matchTime = DateTime.Now;
	}
	void Update()
	{
		if (flag)
		{
			if (Input.GetKey(KeyCode.Return) || Input.GetKey(KeyCode.KeypadEnter))
			{
				if (passwdinput1.gameObject.activeSelf)
				{
					Account();
				}
				else
				{
					Login();
				}
			}
			if (!loginflag)
			{
				var lu = (DateTime.Now - loginTime).TotalMilliseconds;
				if (lu > 2400)
				{
					outText.text = "";
					outText.gameObject.GetComponent<Text>().color = new UnityEngine.Color(0.1f, 0.1f, 0.1f, 1f);
					loginTime = DateTime.Now;
				} else if (lu > 1200)
				{
					outText.gameObject.GetComponent<Text>().color = new UnityEngine.Color(0.1f, 0.1f, 0.1f, 0.5f);
				}
			}
		} else if (ui.todoshow.gameObject.activeSelf)
		{
			var lu = (DateTime.Now - loginTime).TotalMilliseconds;
			if (lu > 2400)
			{
				ui.untodo();
				ui.todotext.GetComponent<Text>().color = new UnityEngine.Color(0, 1, 0.255f, 1);
				ui.titletext.GetComponent<Text>().color = new UnityEngine.Color(0, 1, 0.255f, 1);
				loginTime = DateTime.Now;
			}
			else if (lu > 1200)
			{
				ui.todotext.GetComponent<Text>().color = new UnityEngine.Color(0, 1, 0.255f, 0.5f);
				ui.titletext.GetComponent<Text>().color = new UnityEngine.Color(0, 1, 0.255f, 0.5f);
			}
		}

		//Recv
		if (!matchflag) { return; }

		var totaltime = (int)(DateTime.Now - matchTime).TotalSeconds;
		if (totaltime > 59)
		{
			selectime.text = Convert.ToString(totaltime / 60) + ":" + Convert.ToString(totaltime % 60);
		}
		else
		{
			selectime.text = Convert.ToString(totaltime) + " s";
		}
		totaltime = (int)(DateTime.Now - preTime).TotalMilliseconds;
		if (totaltime < 990)
		{
			return;
		}
		preTime = DateTime.Now;
		byte[] readBuff = new byte[5000];
		int count = socket.Receive(readBuff);
		var res = Response.Parser.ParseFrom(readBuff, 0, count);
		var sta = res.Status;
		selectText.text = res.Msg;
		if (sta == 3)
		{
			string s = "";
			for (int i = 0; i < res.UserList.Count; i++)
			{
				s = s + res.UserList[i].Name + " 、";
			}
			selectText2.text = s;
			s = "";
		} else if (sta == 0)
		{
			room = res.Room;
			match2room();
			matchflag = false;
		}
	}

	// 点击弹窗
	public void Popup()
	{
		popwindows.SetActive(false);
		outText.text = popup.text = "";
	}

	public void End()
	{
		if (flag) socket.Close();
		flag = false;
		Application.Quit();
		popwindows.SetActive(true);
		outinfo("游戏已退出!");
		popup.text = "游戏已退出!";
	}


	public void Login2menu()
	{
		ui.login(false);
		ui.menu(true);
		PlayerPrefs.SetString("name", nameinput.text);
		PlayerPrefs.SetString("passwd", passwdinput.text);
		RandomName();
		uid.text = Convert.ToString(user.Id);
		uName.text = userName.text = user.Name;
		ulevel.text = userLevel.text = "等级：" + Convert.ToString(user.Level);
		userExp.text = "Exp：" + Convert.ToString(user.Experience);
		flag = false;
		loginflag = true;
	}
	public void menu2login()
	{
		ui.menu(false);
		ui.login(true);
		ui.rePressbtnA();
		initlogin();
		flag = true;
		loginflag = false;
	}
	public void menu2select()
	{
		ui.menu(false);
		ui.select(true);
	}
	public void menu2skin()
	{
		ui.menu(false);
		ui.skin(true);
	}
	public void menu2match()
	{
		ui.menu(false);
		ui.match(true);
		gifbool = true;
	}
	public void skin2menu()
	{
		ui.skin(false);
		ui.menu(true);
	}
	public void Select2match()
	{
		ui.select(false);
		ui.match(true);
		gifbool = true;
	}
	public void Select2menu()
	{
		ui.select(false);
		ui.menu(true);
	}
	public void match2room()
	{
		ui.match(false);
		ui.room(true);
		selectText2.text = selectText.text = selectime.text = "";
		playerFlag = true;
		user.Name = nickNameinput.text;
		gifbool = false;
	}
	public void room2settle()
	{
		ui.room(false);
		ui.settle(true);
		playerFlag = false;
	}
	public void settle2menu()
	{
		ui.settle(false);
		ui.menu(true);
	}
	public void initlogin()
	{
		ui.login(true);
		ui.menu(false);
		selectText2.text = selectText.text = selectime.text = "";
		outText.text = popup.text = "";
		userExp.text = userLevel.text = userName.text = "";
		nameinput.text=PlayerPrefs.GetString("name", "");
		passwdinput.text= PlayerPrefs.GetString("passwd", "");
		nickNameinput.text = "";
	}

	public void uinfo()
	{
		if (uName.gameObject.activeSelf)
		{
			uid.text = "用户信息";
			uName.gameObject.SetActive(false);
			ulevel.gameObject.SetActive(false);
		}
		else
		{
			uid.text = "uid:" + Convert.ToString(user.Id);
			uName.gameObject.SetActive(true);
			ulevel.gameObject.SetActive(true);
		}
	}

	void outinfo(string s)
	{
		outText.text = s;
		loginTime = DateTime.Now;
	}
	public void RandomName()
	{
		RNGCryptoServiceProvider csp = new RNGCryptoServiceProvider();
		byte[] byteCsp = new byte[10];
		csp.GetBytes(byteCsp);
		var randName = BitConverter.ToUInt32(byteCsp, 0) % ui.nameList.Length;
		nickNameinput.text = ui.nameList[randName];
	}

	bool checkedstring(string name,string passwd){
		if (name.Length > 12 )
        {
			outinfo("用户名超过12字符");
			return false;
        }
		if (name.Length <3)
		{
			outinfo("用户名少于3字符");
			return false;
		}
		if (passwd.Length > 24)
		{
			outinfo("用户密码超过24字符");
			return false;
		}
		if (passwd.Length < 6)
		{
			outinfo("用户密码少于6字符");
			return false;
		}
		return true;
	}

}