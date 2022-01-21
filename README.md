# DCoB-Scheduler
Distributed Computing on Browser Scheduler

## Concepts & Defines
### What
DCoB is a computing platform powered by web browsers as workers.

DCoB's cluster contains one(or a small group) centralized scheduler nodes and many loose worker nodes.

DCoB is not a low latency / high response / multi-tenant job executor, it's intend to run a small number of big jobs with idle computing power. 

### Why
Since all mainstream web browsers are support WASM, and in most time we use web browser to browse pages, our CPU's idle rate usually very high.

Usage such idle computing power, we have so many things can do.

### Features to be developed:
**worker**
- [x] interrupt
- [ ] really need to keep closing status?
- [x] choose function from function id
- [x] display task status in index.html
- [x] deploy to github pages
- [ ] try slim the WASM file size

**scheduler**
- [ ] apply worker not to burn cpu, but callback to notify worker applied
- [x] fetch miner job difficulty from http request
- [x] provide a web ui to display jobs and tasks
- [x] issue functions (WASM bytecode) to worker
- [ ] support concurrent job execute
- [ ] replace job queue to priority queue
- [ ] deploy to server with public ip
- [ ] store job context to DB or file, then release memory
- [ ] distributed deploy, may need distributed or cascaded worker pool and queue
- [ ] try DAG job