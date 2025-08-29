package core

/*
分析我提供的数据，根据我的意图，按照格式返回数据给我。

设备数据：%s;
指令数据：%s;
返回JSON格式:{
        "alias": "自动化名称",
        "description": "自动化功能描述",
        "triggers": [
            {
                "trigger": "time",
				"at":"22:23:05"
            }
        ],
		"conditions": [
		   {
            "condition": "time",
            "weekday": [
                "mon",
                "tue"
            ]
      	  }
       ],
		"actions": [
            {
                "type": "turn_on",
				"domain":"light",
				"brightness_pct":100;
				"entity_id":"68d419bf3cc1a0a94e1d82fe3c5bbda3",
            }
        ]
    }

数据格式说明：
alias：根据意图，给这个自动化命名，要求简短。
description: 对这个自动化的功能进行描述。
triggers：表示在某个时间点执行，如果我的意图比较模糊例如：下午就帮我选一个下午的时间点赋值到at中。
conditions：表示周期性的weekday有7个值可选，分别是周一(mon)、二(tue)、三(wed)、四(thu)、五(fri)、六(sat)、日(sun)。
actions：表示要控制的设备，通过设备数据和指令数据得到
*/
// 定时周期执行的动作：设备，场景，自动化
