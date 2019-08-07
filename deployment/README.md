# How to bring order to the world

To run everything at once for test purposes, go to `docker-compose` folder and run
```
docker-compose up
```

You can run each component by 
```
docker-compose docker-compose -p <sometoken> run <component>
```
example: Add one more worker:
```.env
docker-compose -p p1 run worker
```

