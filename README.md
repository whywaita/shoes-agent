# shoes-agent

Framework for [myshoes](https://github.com/whywaita/myshoe) provider using agent.

- `agent`: agent for `shoes-agent`. `agent` run on insntance for runner.
- `proto`: file of protoc. It define RPC for `agent`
- `shoes-agent`: package of framework.
- `shoes-agent-mock`: example implementation using `shoes-agent`

## Workflow

```
@startuml
myshoes -> "shoes-agent-xxx" : AddInstance()
"shoes-agent-xxx" -> "Backend API" : Create an instance
"Backend API" --> "shoes-agent-xxx" : Created
"shoes-agent-xxx" -> "Backend API" : ListAgent()
"Backend API" --> "shoes-agent-xxx" : List of agent (cloud ID, status)
"shoes-agent-xxx" -> "shoes-agent-xxx" : Do scheduling an instance from list of agent
"shoes-agent-xxx" -> "agent in instance" : StartRunner()
"agent in instance" ->> "agent in instance" : Execute a script
"agent in instance" --> "shoes-agent-xxx" : Started
"shoes-agent-xxx" --> "myshoes": Created
@enduml
```

![](http://www.plantuml.com/plantuml/png/ZP91QWCn34NtSmengmdK5-YYPDfqCT15QKzWR2KnjULYoU3UlYW9924nP6VCV__xIrwps28rnI7zyJuZWtc1yN0oTeSafhKsmZFCtY_4OidXj1fk5OgzMlU3v67-N1HvAsW5mHA44pbSIqmdwmZwnr8-0ikiYcdreBqIaBTmk8J9nLmzB9idOB5I-LwxZjCc0xiz-Xe3xIwBmhRa1F4ogEDwV4Gue-hxhKlvgaHOjDMDHj4U-zxGLHqxi2lXL-xZdK8Qt9cy4gS_CfvNj4RoDSL_)
