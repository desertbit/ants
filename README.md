# ANTS - Let the ants handle your serial communication.
**WARNING:** This protocol may change slightly without version mutation within the next few days.

This protocol is an asynchronous binary data protocol for communication over e.g. serial ports. It takes care about synchronization, checksum validation and error handling. The end user is able to transmit data chunks without worrying about the communication layer. Other protocols can be easily build up based on the ANTS protocol.

**The protocol can be found [here](protocol.md)**

# Support
Feel free to contribute to this project.

# TODO
- Protocol addition: check the PMSN for invalidity to ignore duplicate data messages.
- Implement the thread-safe Golang libraries.
- Implement an automatic test program to test new clients for a valid protocol implementation.
