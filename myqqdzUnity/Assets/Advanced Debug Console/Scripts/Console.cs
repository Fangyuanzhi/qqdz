//https://polygonplanet.com/
//Copyright © 2016 Polygon Planet. All rights reserved.
//This source file is protected under UNITYS Asset Store Terms of Service and EULA which can be viewed here https://unity3d.com/legal/as_terms

#pragma warning disable 0168 //Variable declared, but not used.
#pragma warning disable 0219 //Variable assigned, but not used.
#pragma warning disable 0414 //Private field assigned, but not used.

using System;
using System.Collections;
using System.Collections.Generic;
using System.IO;
using UnityEngine;
using UnityEngine.UI;

public class Console : MonoBehaviour
{
    [Header("UI")]
    public GameObject terminalUI;
    public Text terminalText;

    [Header("Info UI")]
    public Text fpsText;
    public Text resText;
    public Text cpuText;
    public Text gpuText;
    public Text ramText;
    public Text spamText;
    public Text restrictLogsText;
    public Text logtoFileText;

    [Header("Log Colors")]
    public string assertColor;
    public string errorColor;
    public string exceptionColor;
    public string logColor;
    public string warningColor;

    [Header("Variables")]
    public int maxLogs;
    public bool spam;
    public bool restrictLogs;
    public bool logToFile;

    //Variables
    private KeyCode toggleKey;
    private bool visible;
    private List<Log> logs;
    private string fullLog;
    private float fps;

    private void Awake()
    {
        DontDestroyOnLoad(gameObject);
        toggleKey = KeyCode.Backspace;
        logs = new List<Log>();
        StartCoroutine(CalculateFPS());
    }

    private void OnEnable()
    {
        Application.logMessageReceived += HandleLog;
    }

    private void OnDisable()
    {
        Application.logMessageReceived -= HandleLog;
    }

    private void Update()
    {
        if (Input.GetKeyDown(toggleKey))
        {
            visible = !visible;
        }

        if (visible == false)
        {
            terminalUI.SetActive(false);
        }
        else
        {
            terminalUI.SetActive(true);
        }

        if (spam == true)
        {
            Debug.Log("<color=red>Spam</color>");
        }
    }

    private void OnGUI()
    {
        ShowLog();
        if (logToFile == true)
        {
            File.WriteAllText(Application.persistentDataPath + "/log.txt", fullLog);
        }

        fpsText.text = "FPS: " + Math.Round(fps, 0).ToString();
        resText.text = "RES: " + Screen.currentResolution.ToString().Substring(0, Screen.currentResolution.ToString().IndexOf('@'));
        cpuText.text = "CPU: " + SystemInfo.processorFrequency;
        gpuText.text = "GPU: " + SystemInfo.graphicsMemorySize + " MB";
        ramText.text = "RAM: " + SystemInfo.systemMemorySize + " MB";

        if (spam == true)
        {
            spamText.text = "Spam: ON";
        }
        else
        {
            spamText.text = "Spam: OFF";
        }
        if (restrictLogs == true)
        {
            restrictLogsText.text = "Restrict Logs: ON";
        }
        else
        {
            restrictLogsText.text = "Restrict Logs: OFF";
        }
        if (logToFile == true)
        {
            logtoFileText.text = "Log to File: ON";
        }
        else
        {
            logtoFileText.text = "Log to File: OFF";
        }
    }

    private void HandleLog(string message, string stackTrace, LogType logType)
    {
        Log log = new Log()
        {
            message = message,
            stackTrace = stackTrace,
            logType = logType
        };
        logs.Add(log);  
        TrimExcessLogs();
    }

    private void TrimExcessLogs()
    {
        if (!restrictLogs)
        {
            return;
        }
        else
        {
            var amountToRemove = Mathf.Max(logs.Count - maxLogs, 0);
            if (amountToRemove == 0)
            {
                return;
            }
            else
            {
                logs.RemoveRange(0, amountToRemove);
            }
        }
    }

    public void Clear()
    {
        logs.Clear();
    }

    private void ShowLog()
    {
        fullLog = "";
        int i = new int();
        foreach (Log log in logs)
        {
            if (logs[i].logType == LogType.Assert)
            {
                fullLog = fullLog + "<color=" + assertColor + ">" + logs[i].message + "</color>" + "\r\n";
            }
            else if (logs[i].logType == LogType.Error)
            {
                fullLog = fullLog + "<color=" + errorColor + ">" + logs[i].message + "</color>" + "\r\n";
            }
            else if (logs[i].logType == LogType.Exception)
            {
                fullLog = fullLog + "<color=" + exceptionColor + ">" + logs[i].message + "</color>" + "\r\n";
            }
            else if (logs[i].logType == LogType.Log)
            {
                fullLog = fullLog + "<color=" + logColor + ">" + logs[i].message + "</color>" + "\r\n";
            }
            else if (logs[i].logType == LogType.Warning)
            {
                fullLog = fullLog + "<color=" + warningColor + ">" + logs[i].message + "</color>" + "\r\n";
            }
            i++;
        }
        terminalText.text = fullLog;
    }

    private IEnumerator CalculateFPS()
    {
        while (true)
        {
            fps = 1 / Time.deltaTime;
            yield return new WaitForSeconds(1);
        }
    }

    public void ChangeSpam()
    {
        spam = !spam;
    }

    public void RestrictLogs()
    {
        restrictLogs = !restrictLogs;
    }

    public void LogToFile()
    {
        logToFile = !logToFile;
    }
}

[Serializable]
public class Log
{
    public string message;
    public string stackTrace;
    public LogType logType;
}