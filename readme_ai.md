## 需求
1. 能发送请求和接受回调的一个测试服务
2. 通过命令行运行测试用例，创建测试文件夹，文件夹名称为test-日期字符串，测试请求的参数从csv里面读取，示例csv表头
app_id,role_desc,scenes,style
3. 发送测试请求后需要在对应测试文件夹内创建文件夹，文件夹的名称为对应的task_id，并将请求信息（role_desc和scenes）记录到文件夹下的desc.txt文件
4. 收到回调结果后，需要将图片下载下来，根据图片的index和scenes中的元素对应，下载的图片放到对应task_id文件夹下，图片名称为对应的scene名称
5. 提供简单的web界面，能够展示测试文件夹列表、task文件列表和图片
6. 用golang实现
## 请求示例
curl --location --request POST 'http://192.168.11.4:7865/comfyui/scene_text2img' \
--header 'Content-Type: application/json' \
--data-raw '{
    "app_id":11,
    "scenes":["friends，「天生一对」"],
    "role_desc": "女 白色裙子 23岁 ",
    "callback":"localhost:9983/callback",
    "style":"real"
}'

## 响应示例
{
    "code": 0,
    "msg": "Success",
    "task_id": "scene_text2img___098971d7-4703-11f0-8d7f-08bfb88182a2",
    "cost_time": 88,
    "queue_count": 0
}

## 回调示例
{
  "task_id":"",
  "success":true,
  "msg":"xxx",
  "data":{
    "imgs":[
      {
        "url":"xxxx",
        "index":0
      },
      {
        "url":"xxxx",
        "index":1
      },
    ]
  }

}