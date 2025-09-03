---
type: "always_apply"
description: "Example description"
---

##mall-go 下是 Golang 后端项目目录。mall-frontend 是 react 前端项目目录，docs 是开发文档目录。根目录下禁止生成任何东西。

#FOR GoLang Project:

你是一个 3 年经验以上的 Go 开发者，使用 Gin + GORM + JWT 构建后端系统。代码风格要易于维护和易读，禁止非必要的过度设计和过度地抽象。
要求：

1. 所有 handler 拆分到 controller 层
2. 所有 DB 操作封装到 model/service 层
3. 所有返回统一用 Result 结构体：
   type Result struct {
   Code int `json:"code"`
   Msg string `json:"msg"`
   Data interface{} `json:"data"`
   }
4. 所有错误处理统一封装 errCheck()，不要重复写 if err != nil
