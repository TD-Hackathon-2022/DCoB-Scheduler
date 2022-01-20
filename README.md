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

Task abstraction should contain:
- task id
- context: stage initial / intermediate / final data
- function: stateful or stateless procedure that can take input and return output.
  - func id 
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
```
    +---------------------+                                                                            
  |-|                     |                                                                            
|-- |                     |                                                                            
| | |  Job                |                                    Task Q                                  
| | |        |------------|     Tasks   +-------------------------------------------------+            
| | |        |            |         \   |    |    |    |    |    |    |    |    |    |    |  ----|     
| | |        |            |   --------  |    |    |    |    |    |    |    |    |    |    |      |     
| | |        | Job        |         /   +-------------------------------------------------+      |     
| | |        | Spliterator|                                                                      |     
| | |        |            |                                                                      |     
| | +--------|---------|--+                                                                      |     
| -------------------|-|                                       +-----------------+               |     
---------------------|                                    \    |                 |    /          |     
                                             |--------------   |     Decider     |   -------------     
                                             |            /    |                 |    \                
                                             |                 +-----------------+                     
                                             |                           |                             
                                             |                           |                             
                                             |                          \|/                            
                                                                         |                             
                                   +-----------------+         +-----------------+                     
                                   |                 |         |                 |                     
                                   |   Worker Pool   |         |   Distributor   |                     
                                   |                 |         |                 |                     
                                   +-----------------+         +-----------------+                     
                                                                                                       
                                         Status                     Cmd     Task                       
                                           -                             |                             
                                          /|\                            |                             
                                           |                             |                             
                                 ----------|-----------------------------|---------------              
                                           |                             |                             
                                           |                             |                             
                                           |                             |                             
                                           |                            \|/                            
                                           |           Worker            -                             
```

1. Job: contains job meta and context
  - Job Spilterator: spilt job to tasks and can be called iteratively to get one task
2. Task Q: 
  - queue
3. Worker Pool:
  - manage worker's lifecycle
  - monitor workers status
4. Decider:
  - pop task from task q
  - apply some workers from worker pool
  - decide how to assign tasks to workers (by some policy)
5. Distributer:
  - distribute task to worker
  - send cmd (start / stop / retry / interrupt) to worker

### Worker
```
                                           
                                           
    Status                  Cmd Task       
      ^                        |           
      |                        |           
      |                        |           
------|------------------------|-----------
      |                        |           
      |                        v           
 +-------------------------------------+   
 |                                     |   
 |                                     |   
 |             Connector               |   
 |                                     |   
 +-------------------------------|-----+   
        ^                        |         
        |                        |         
 +------|-----+           +------v-----+   
 |            |           |            |   
 |  Monitor   <-----------| Controller |   
 |            |           |            |   
 +------------+           +------^-----+   
                                 |         
                                 |         
                                 |         
                                 |         
                          +------v-----+   
                          |            |   
                          |  Executor  |   
                          |            |   
                          +------------+   
```

1. Connector:

- maintain connection status
- register, cancellation

2. Controller:

- assign task to executor
- start / stop / retry / interrupt
- gathering status / result

3. Executor:

- run task

4. Monitor:

- heartbeat / lease
- task status
- task result

### Demo

This demo will try to mine simple virtual coins

#### reference

> https://github.com/dockersamples/dockercoins

### Features to be developed:
**worker**
- [x] interrupt
- [ ] really need to keep closing status?
- [x] choose function from function id
- [x] display task status in index.html
- [ ] deploy to github pages
- [ ] try slim the WASM file size

**scheduler**
- [ ] apply worker not to burn cpu, but callback to notify worker applied
- [x] fetch miner job difficulty from http request
- [ ] provide a web ui to display jobs and tasks
- [ ] issue functions (WASM bytecode) to worker
- [ ] support concurrent job execute
- [ ] replace job queue to priority queue
- [ ] deploy to server with public ip
- [ ] store job context to DB or file, then release memory
- [ ] distributed deploy, may need distributed or cascaded worker pool and queue
- [ ] try DAG job