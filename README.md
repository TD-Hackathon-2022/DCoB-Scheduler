# DCoB-Scheduler
Distributed Computing on Browser Scheduler

## Conceps & Defines
### What
DCoB is a computing platform powered by web browsers as workers.

DCoB's cluster contains one(or a small group) centralized scheduler nodes and many loose worker nodes.

DCoB is not a low latency / high responsee / multi-tenant job executor, it's intend to runing a smal number of big jobs with idle computing power. 

### Why
Since all mainstream web browers are support WASM, and in most of time we use web brower to browse pages, our CPU's idle rate usually very high.

Usage such idle computing power, we have so many things can do.

### Job
We define job as some kind of computing process.

The qualified job should:
- CPU intensive, not IO intensive
- Can be easily spilt into many small independent or weakly relevant tasks, which can distribute to workers. 

### Task
Task is one piece of job, every task should execute on some workers.

Task abstraction should contains:
- context: stage initial / intermediate / final data
- function: stateful or stateless procedure that can take input and return output.
- control handler: responsible for start / stop / retry / interrupt

### Resource
Workers (browsers) have those characteristics:
- Unstable: to avoid our task affect the normal user browse, worker's process priority is relatively lower.
- Volatile: task can be killed with user close their browser.
- Homogeneous: no matter user's operating system, CPU architecture, browser type, all tasks running on WASM VM.

## Architecture
```
                                 TASK TO DAG
                                      |
                                      |
                                      |
                                      |
                                      |
                                      |
                          +-----------v------------+
                          |                        |
                          |      Scheduler         |
                          |                        |
                          +------------------------+
                          |                        |
                          |     Worker Pool        |
                          |                        |
                          +------------------------+
                          |                        |
                  +------->    Connection Registry <----------------+
                  |       +----------------^-------+                |
                  |                        |                        |
------------------+------------------------+------------------------+---------------
                  |                        |                        |
                  |                        |                        |
                  |                        |                        |
                  |                        |                        |
                  |                        |                        |
                  |                        |                        |
        +---------+---------+      +-------+-------+        +-------+------+
        |                   |      |               |        |              |
        |    Task1RunTime   |      |   Task2RunTime|        |              |
        +-------------------+      +---------------+        |              |
        |                   |      |               |        |              |
        |      WASM         |      |    ...        |        |   ...        |
        +-------------------+      +---------------+        |              |
        |     Nodejs        |      |     ...       |        |              |
        +-------------------+      +---------------+        |              |
        |      V8           |      |     ...       |        |              |
        +-------------------+      +---------------+        +--------------+
```

## Modules Design
### Scheduler

### Worker


