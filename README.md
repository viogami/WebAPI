# Viogami`s WebAPI

个人web后端，提供多种小玩意的api，目前只提交了可公开的部分.

目前搭载功能：

* [p5r预告信生成功能api](https://github.com/viogami/WebAPI/tree/master/core/p5cc)，通过动态路由 `../p5cc/:text` 生成包含该text的p5r风格预告信，一起偷心吧。
  `<br>`（发帖宣传在小黑盒：[https://www.xiaoheihe.cn/app/bbs/link/146347423](https://www.xiaoheihe.cn/app/bbs/link/146347423)）
* [wxapi功能](https://github.com/viogami/WebAPI/tree/master/core/wxapi)，集成控制微信公众号后台的收发消息，目前接口搭载了基于gpt的消息回复.

* [AI回复功能](https://github.com/viogami/WebAPI/tree/master/core/AI)，集成openai和deepseek的api，提供多种ai功能。

架构：

* 基于golang的gin框架，读取配置文件 `config.yaml`
* core是核心业务逻辑层
* 服务层封装了路由和一系列handler，用于监听请求
* middleware为中间件，目前只搭载了限流。短时间同一ip频繁请求会被永久封禁。
