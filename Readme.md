# Mongo Driver Test
This is a project to test the behavior of the golang MongoDB driver under high load conditions

## How to run
In order to run the project all you need to do is run the main.go file, it will start a gin server listening on port 8090, it serves 2 paths with 2 different methods:

GET    /api/v1/health
POST   /api/v1/stages/

Once the server is up, you can start the test by sending a POST method to the /api/v1/stages/ URI, with the test payload in the body of the request (see the payload section)

## Payload

The /api/v1/stages/ will receive a POST call and will evaluate the payload sent in the body to prepare the test and run it, the payload is divided in 2 sections, each with its own parameters, they are:

### db_config
####   db_name: 
The name of the MongoDB database
####   collection_name: 
The name of the collection
####   conn_string: 
The full connection string to the MongoDB database 
####   min_pool_size: 
The minimum connection pool size
####   max_pool_size: 
The maximum connection pool size
####   idle_timeout: 
The idle timeout 
####   socket_timeout: 
The socket timeout

### stage_config:
####   workers_count: 
The number of initial workers for the test
####   workers_to_add: 
The number of workers to add at each step of the test
####   increment_load: 
The number of times(steps) the test will increment the load####   producers_count: The number of producers at each step of the test
####   msg_by_sec: 
The number pf messages per sec to be sent
####   time_to_sleep_secs: 
The time elapsed between each step of the test
####   time_to_finish_secs: 
The time to wait at the end of test before finishing
####   query_timeout_ms: 
The timeout parameter passed to each query on the Find() method
####   batch_size: 
The batch size parameter passed to each query on the Find() method, 0 for no batch size (it will use the default)
####   collection_size: 
The number of objects to be created in the database for the test
####   document_size_kb: 
The size in Kb of each object to be created in the database for the test (this is aproximate)

## Example payload

{
	"db_config":{
		"db_name":"stores",
		"collection_name": "stores",
		"conn_string":"mongodb://test:test@localhost:27017/stores",
		"min_pool_size":30,
		"max_pool_size":100,
		"idle_timeout":60,
		"socket_timeout":50
	},
	"stage_config":{
		"workers_count":10,
		"workers_to_add":45,
		"increment_load":2,
		"producers_count":50,
		"msg_by_sec": 20,
		"time_to_sleep_secs":45,
		"time_to_finish_secs":10,
		"query_timeout_ms": 50,
		"batch_size": 0,
		"collection_size": 10000,
		"document_size_kb": 1
	}
}

In this example, the test will connect to a local mongoDB instance using the user "test" with password "test" and the "stores" database.

The test will first create 10000 documents of 1 Kb in size (each one), and then start running, it will initially create 50 producers and 10 workers, sending 20 queries per sec, and setting a timeout of 50 ms and no batch size(default value of 0) for each query.

It will wait for 45 sec before adding 45 additional workers (for a total of 55 running), and then after another 45 secs adding 45 more for a total of 100

Then after another 45 secs, it will stop the workers and wait for 10 secs before finishing the tests and presenting the final results 
 