# Viogami`s WebAPI

个人web后端，提供多种小玩意的api，目前只提交了可公开的部分。

基于golang，使用gin框架，读取配置文件 `config.yaml`,

目前搭载功能：

* p5r预告信生成功能api，通过动态路由 `../p5cc/:text` 生成包含该text的p5r风格预告信，一起偷心吧。
