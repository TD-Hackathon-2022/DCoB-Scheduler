## Message Exchange Between Scheduler <=> Worker

### Basic Structure
1. Normal Msg
As normal msg, we simply use json with two fields:
- CMD
  - type: [Register(0) | Close(1) | Status(2) | Assign(3) | Interrupt(4)]
- PAYLOAD

*example msg:*
```json
{
  "CMD": 0,
  "PAYLOAD": {...payload detail...}
}
```

2. Heartbeat
As heartbeat msg, we use websocket ping/pong to make it happen.
- Ping: send from Worker, with fixed interval
- Pong: send from Scheduler

### Message Details
#### Register
```json
{
  "CMD": 0,
  "PAYLOAD": null
}
```

#### Close
```json
{
  "CMD": 1,
  "PAYLOAD": null
}
```

#### Status
```json
{
  "CMD": 2,
  "PAYLOAD": {
    "workerStatus": 1,
    "taskId": "task-id",
    "taskStatus": 0,
    "execResult": "...BASE64 ENCODED..."
  }
}
```

workerStatus
- 0: idle
- 1: busy
- 2: closing (prepare to close, no further task accept)

taskStatus
- 0: running
- 1: finished
- 2: error
- 3: interrupted

#### Assign
```json
{
  "CMD": 3,
  "PAYLOAD": {
    "taskId": "task-id",
    "data": "...BASE64 ENCODED...",
    "funcId": "func-id"
  }
}
```

Currently, UDF is not supported yet, so we use `funcId` to point built-in functions.

#### Interrupt
```json
{
  "CMD": 4,
  "PAYLOAD": {
    "taskId": "task-id",
  }
}
```