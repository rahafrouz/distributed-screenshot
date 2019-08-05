# S4 Scalable ScreenShot Service
This is a possible implementation for a service for taking screenshot of a webpage.

## S4 in one sentence 
The input is a number of URLs, and the output is a number of screenshots.

## Project Structure
There are two alternative solutions that we can choose for this service. 
- With message Broker
- Without message Broker

There is a trade-off between complexity and cost. 

The full implementation is the approach with message-broker, due to ease of extension for future features. 

### Broker-based Approach
The `deployment`, `worker`,`dispatcher` are involved with the broker-based approach. 
This is the diagram of the system. 

`diagram of broker and worker and dispatcher`

###### Pros 
- It is simpler to implement. Less concurrent code means less hard-to-find bugs.
- A `worker` instance could be in an internal network to do its job. 
In this case that it needs to connect to the internet for taking screenshot,
 it could be hidden a private network behind a NAT. The NAT doesn't allow connections to be initiated from outside of the network.
- If a `worker` stops working, the `broker` would detect if faster, and the message could be handed over to another worker by the RabbitMQ broker. 
This nice feature can be handled by proper configuration, without the need to re-invent the wheel.
- If someone else wants to extend the work, they read the worker code that is about the logic, not the code about how concurrency is handled, 
which may save time and energy in the long-run, and a more maintainable code.
###### Cons
- Broker is the single point of failure. The whole solution could be mirrored as fail-over service, but cost, complexity?
- Possible vulnarabilities, immaturities in the RMQ protocol. 

### Broker-less Approach
TODO: not complete.
###### Pros
This way, a single micro-service is developed in go. This micro-service is responbile for the quality of the service. The micro-service, should know its limits,
and if the micro-service's resources are getting saturated, it should not get more requests, since it will increase latency of the requests. 
- The micro-service would have a pool of go-routines that handle screenshot taking in parallel. This parameter depends on the CPU, Memory capability of the OS, since we use chrome-headless, it should be measured over time, to fine-tune this parameter.
- The micro-service would have a pool of go-routines that handle S3 operations, to prevent exhausting file descriptors. 
(We cannot have more than a number of open socket connections.)

