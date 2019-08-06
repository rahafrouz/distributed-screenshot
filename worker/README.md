# The System Architecture

There are different engines for taking screenshot. 
Here we use the package called `gowitness`. 
It uses `chrome-headless` as the internal engine. 

Other alternatives are:
- **eyewitness**: It is an option. It _seems_ they use firefox engine.  
- **chrome-headless**: For this package, it seems the emphasis is on the debug protocol for comprehensive commands. 
It seems promising, however, using the command line, there was no way to change the output of the file (It is screenshot.png). _There is a way, but I could not find it._ TODO For Future.   
- **PhantomJS**: It utilized the debug protocol of `chrome headless`. It seems a more mature option.

The option of gowintess, is not the best, however, it does the job for current requirement. 
To replace it in the future, just the file `worker.go` should be modified. 

