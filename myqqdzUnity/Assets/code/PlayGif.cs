using System;
using System.Collections.Generic;
using System.Drawing;
using System.Drawing.Imaging;
using System.IO;
using UnityEngine;

public class PlayGif : MonoBehaviour
{

    public UnityEngine.UI.Image Im;
    SpriteRenderer sp;
    public string gifName = "source/front/ba54611d2d6f425eb7f557601f835bd9";
    public string gifplay = "source/front/8248585c2ae34b87b19b8ea402661d56";
    public GameObject[] Ims;
    [SerializeField]
    private float fps = 30f;
    private List<Texture2D> tex2DList = new List<Texture2D>();
    private List<Texture2D> play2DList = new List<Texture2D>();
    private float time,sx,sy;
    public int playgif=0;
    public Echo echogif;
    void Start()
    {
        Image image = Image.FromFile(Application.dataPath +"/"+ gifName + ".gif");
        tex2DList = MyGif(image);
        image=Image.FromFile(Application.dataPath + "/" + gifplay + ".gif");
        play2DList = MyGif(image);
        Debug.Log("读入的图片长度："+play2DList.Count);
        sx = 600f/image.Width;
        sy = 600f/image.Height;
        sp =echogif.player.transform.Find("skinSprite").GetComponent<SpriteRenderer>();
    }

    // Update is called once per frame
    void Update()
    {
       
        if (echogif.gifbool&&tex2DList.Count > 0)
        {
            time += Time.deltaTime;
            int index = (int)(time * fps) % tex2DList.Count;
            if (Im != null)
            {
                Im.sprite = Sprite.Create(tex2DList[index], new Rect(0, 0, tex2DList[index].width, tex2DList[index].height), new Vector2(0.5f, 0.5f));
                sp.transform.localScale = new Vector3(sx, sy, 1);
            }
        }
        if (echogif.playerFlag && playgif > 0&& play2DList.Count > 1)
        {
            time += Time.deltaTime;
            int index = (int)(time * fps) % play2DList.Count;
            if (sp != null)
            {
                sp.sprite = Sprite.Create(play2DList[index], new Rect(0, 0, play2DList[index].width, play2DList[index].height), new Vector2(0.5f, 0.5f));
                sp.transform.localScale = new Vector3(sx, sy, sp.transform.localScale.z);
            }            
            playgif--;
        }
        else if(echogif.playerFlag && playgif==0&&sp.sprite != null)
        {
            var id = echogif.ui.skinId;
            sp.sprite= echogif.ui.imgMap[id];
            sp.transform.localScale = new Vector3(600f/echogif.ui.skinMap[id].Width, 600f / echogif.ui.skinMap[id].Height, 1);
        }
    }

    public List<Texture2D> MyGif(System.Drawing.Image image)
    {

        List<Texture2D> tex = new List<Texture2D>();
        if (image != null)
        {

            //Debug.Log("图片张数：" + image.FrameDimensionsList.Length);
            FrameDimension frame = new FrameDimension(image.FrameDimensionsList[0]);
            int framCount = image.GetFrameCount(frame);//获取维度帧数
            for (int i = 0; i < framCount; ++i)
            {

                image.SelectActiveFrame(frame, i);
                Bitmap framBitmap = new Bitmap(image.Width, image.Height);
                using (System.Drawing.Graphics graphic = System.Drawing.Graphics.FromImage(framBitmap))
                {
                    graphic.DrawImage(image, Point.Empty);
                }
                Texture2D frameTexture2D = new Texture2D(framBitmap.Width, framBitmap.Height, TextureFormat.ARGB32, true);
                frameTexture2D.LoadImage(Bitmap2Byte(framBitmap));
                tex.Add(frameTexture2D);
            }
        }
        return tex;
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

    public void playskin(GameObject gb)
    {
        playgif = 150;
        sp = gb.transform.Find("skinSprite").GetComponent<SpriteRenderer>();
    }
}
