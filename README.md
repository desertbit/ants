# ANTS - Let the ants handle your serial communication.
**WARNING:** This protocol is not ready for production use and is still in active development.

This protocol is an asynchronous binary data protocol for communication over e.g. serial ports. It takes care about synchronization, checksum validation and error handling. The end user is able to transmit data chunks without worrying about the communication layer. Other protocols can be easily build up based on the ANTS protocol.

**The protocol can be found [here](protocol.md)**

# Support
Feel free to contribute to this project.

# TODO
- Protocol addition: check the PMSN for invalidity to ignore duplicate data messages.
- Implement the thread-safe Golang libraries.
- Implement an automatic test program to test new clients for a valid protocol implementation.
- Test tool: create a test case to check if the peer DLE escaping was implemented right.
