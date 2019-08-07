# The Dispatcher
This is a very simple implementation of the dispatcher.
It takes data from a file containing a list of URLs (One per line). 
This dispatcher could have RESTful interface, and get requests from network. 
## Scale it up
It is an stateless element, and it is designed to be replicated. It simply connects to broker, and start sending jobs, and receiving responses. It could be scaled up.
Possible strategy could be to scale it up using ELB, when network IO/CPU usage of the underlying system excceeds a threshold.

## Usage 
`docker-compose up` to run the dispatcher. Look at the compose file, and make sure you have a file named `input.data` in this directory.
