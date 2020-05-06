# Mongo Driver Test
Project's objetive is reproduce the faulty behaviour of some microservices at Rappi.

This behaviour comprises three components:
* Microservice' implementation: Covered by the test implementation, doing some random querys
* MongoDB' driver: Implementing connection's pool logic and all the necessary stuff to query datasource
* Db engine: Providing data's storage, access's security and db operations (external component) 

## Some background
MongoDB's connection pool(mcp since now) work with logical connections.
When started, mcp creates logical connections until reach the configuration parameter 
_MinPoolSize_ but delay the physical connection until some process requires
a connection from the mcp.  
When a process requires a connection from mpc, the mpc's logic verifies the state of the logical connection
and if not connected then proceed to do the physical connection to database's engine and when the 
physical connection is done the mcp mark the connection as 'in use' and provides this connection to the 
requester process.
During the process of searching a connection from the pool, the mcp validates each candidate connection's idle time 
agains the value configured by _MaxConnIdleTime_ and when this max idle time was reached, then the connection 
is replaced by a new one, and the physical connection process starts on this new logical connection.

## Conjeture
Microservices present some failures when getting idle logic connections from mcp and new connections
can't connect to database engine.
This behaviour probably signal a problem when trigger a bunch of authentication process to database engine.


## Running test's scenarios
This service limit the running scenarios to one at time.  
All the process, including connection pool, test data, query executions, and so on,
lives at stage execution and no one survives between executions.
 