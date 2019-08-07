# The System Architecture

## Run it
To run, first edit .env file and enter the credentials. Then enter `docker-compose up`. Voila!

## Alternatives
There are different engines for taking screenshot. 
Here we use the package called `chromedp`. 
It uses `chrome-headless` as the internal engine. 

Other alternatives are:
- **eyewitness**: It is an option. It _seems_ they use firefox engine.  
- **chrome-headless**: For this package, it seems the emphasis is on the debug protocol for comprehensive commands. 
It seems promising, however, using the command line, there was no way to change the output of the file (It is screenshot.png).
 _There is a way, but I could not find it._ TODO For Future.   
- **PhantomJS**: It utilized the debug protocol of `chrome headless`. It seems a more mature option.
- **ChromeDP**: This is the current implementation. The API is from chrome developer team, first-hand, and does not have any intermediary such as gowitness (Which has problems in concurrency) It has Golang API to call to take screenshot. 
Nice to use as a library, while for others such as `gowitness`, you would have to call OS-level command to run it, which gives you less flexibility. 
This engine can have extra options such as resolution, vieport, ... for future development.
#### Extentding the functionality
Right now, the `chromedp` engine is implemented. To use another engine, look at `common/utils.go`. Create a data structure that agrees to ` ScreenshotHanlder ` interface.

# Build Instructions.
To build the worker, refer to `build` file. Most of the options are feeded into the worker from the environment variables. The `docker-compose` is used for deployent. 

*Currently, an executalbe for linux architecture is generated, and added to a container. The container contains chrome.*