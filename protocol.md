# ANTS Protocol Version 1.0.0
## 1. General
**ANTS - Let the ants handle your serial communication.**

This protocol is an asynchronous binary data protocol for communication over e.g. serial ports. It takes care about synchronization, checksum validation and error handling. The end user is able to transmit data chunks without worrying about the communication layer. Other protocols can be easily build up based on the ANTS protocol.

### 1.1 Authors
- Roland Singer - roland.singer[at]desertbit.com

### 1.2 License
This protocol is licensed under the [GNU General Public License Version 3](http://www.gnu.org/licenses/).

### 1.3 Versioning
The protocol version follows the [Semantic Versioning Scheme](http://semver.org/).

Given a version number **MAJOR.MINOR.PATCH**, increment the:

1. **MAJOR** version when you make incompatible API changes,
2. **MINOR** version when you add functionality in a backwards-compatible manner, and
3. **PATCH** version when you make backwards-compatible bug fixes.

### 1.4 Peers
There are no Master or Slave peers. Both communication peers have the same logic and do not differentiate. However it is possible to implement a Master/Slave protocol based on this protocol (Check the Master/Slave Protocol section for more information).

## 2. Data Link Control
This protocol uses the Data Link Escape (DLE) character to differentiate between control characters and the binary data transmission.

The **DLE** character has the constant value of **0x10**.

### 2.1 Control Characters
Control characters are preceded with the DLE character.

NAME | VALUE | DESCRIPTION
---- | ----- | --------------------
STX  | 0x02  | Start of text
ETX  | 0x03  | End of text
ACK  | 0x06  | Acknowledge
NAK  | 0x15  | Negative Acknowledge

### 2.2 Data Encoding
Whenever the DLE character is encountered in the data, it is sent twice to prevent the byte that follows from being interpreted as a control character.

## 3. Message Format
There are two types of messages. Data messages transmit data chunks and control messages are responsible for the flow control.

### 3.1 Data Messages
#### Format
Data messages are defined as below:

STX    | Message Sequence Number | Append Data Flag | Binary Data Body   | CRC 16/32 Checksum | ETX
------ | ----------------------- | ---------------- | ------------------ | ------------------ | ------
1 Byte | 1 Byte                  | 1 Byte           | Maximum 1024 Bytes | 2/4 Bytes          | 1 Byte

### 3.2 Control Messages
Control messages have a higher priority and therefore a precedence over data messages. They are always send as soon as possible, even if there are data messages available in the send queue.

#### 3.2.1 Acknowledge Control Message
The acknowledge control message tells the other peer that the data message with the specific message sequence number was received, validated and successfully encoded.

##### Format

ACK    | Message Sequence Number | CRC-16 Checksum | ETX
------ | ----------------------- | --------------- | ------
1 Byte | 1 Byte                  | 2 Bytes         | 1 Byte

#### 3.2.2 Negative Acknowledge Control Message
The negative acknowledge control message tells the other peer that the previously received data message with the specific message sequence number was not received successfully. The sender peer has to resend the data message again. If the message sequence number is unknown, due to a corrupted data message, then use the unknown message sequence number (UMSN). The sender peer knows anyway that the previously send message data was the corrupted one.

##### Format

NAK    | Message Sequence Number | CRC-16 Checksum | ETX
------ | ----------------------- | --------------- | ------
1 Byte | 1 Byte                  | 2 Bytes         | 1 Byte

## 4. CRC - Cyclic redundancy check
A cyclic redundancy check (CRC) is an error-detecting code commonly used in digital networks and storage devices to detect accidental changes to raw data. Blocks of data entering these systems get a short check value attached, based on the remainder of a polynomial division of their contents. On retrieval the calculation is repeated, and corrective action can be taken against presumed data corruption if the check values do not match.

Computation of a cyclic redundancy check is derived from the mathematics of polynomial division, modulo two. In practice, it resembles long division of the binary message string, with a fixed number of zeroes appended, by the "generator polynomial" string except that exclusive OR operations replace subtractions.

This protocol uses **16-bit** CRC checksums for control messages and either **16-bit** or **32-bit** CRC checksums for data messages. Please refer to [Wikipedia](https://en.wikipedia.org/wiki/Computation_of_cyclic_redundancy_checks) for more information. The CRC is computed over the complete message body including the STX and except the CRC and ETX characters.

### 4.1 Polynomial

CRC    | NAME    | VALUE      | DESCRIPTION
------ | ------- | ---------- | ----------------------------------------------------------------
CRC-16 | CCITT   | 0x8408     | Used by X.25, V.41, HDLC FCS, XMODEM, Bluetooth, PACTOR, SD, ...
CRC-32 | Koopman | 0xeb31d82e | Better error detection characteristics than IEEE

## 5. Message Sequencing
The sequence number is a **1 byte decimal** value which cycles from **1 to 255**. As soon as the highest sequence number with a value of **255** is reached, then the cycle continues with the next value **1**.

VALUE | DESCRIPTION
----- | --------------------------------------
0     | Unknown message sequence number (UMSN)
1-255 | Normal message sequence number

### 5.1 Message Sequence Number (MSN)
Both peers have a separate message sequence number (MSN), which is separately incremented. An initial message sequence number has the value 1. Message sequence numbers are incremented by the owner peer before each data message transmissions. It does not matter, if the data message transmissions was successful, the MSN is always incremented.

The MSN should be echoed back from the communication peer after a data message transmissions. It has to match the MSN of the data message.

### 5.2 Peer Message Sequence Number (PMSN)
Peer message sequence numbers are the received message sequence numbers (MSN) from the communication peer. They are not incremented by the receiving peer. A received PMSN is valid if it is a decimal number within the defined range of (0-255). PMSNs are send back to the sending peer within a control message. Other validations are not made by the receiving peer.

_Hint: the MSN is called PMSN on the receiver peer._

### 5.3 Samples
#### Successful message transmission
Peer 1 sends a data message to peer 2. The MSN is incremented and attached to the data message. Peer 2 successfully validates, encodes and receives the data message from peer 1 and sends an Acknowledge Control Message to peer 1 to signalize a successful reception. The PMSN is included into this control message with the same value as received from peer 1.

```
PEER 1 (MSN=1)     // Initial MSN value

PEER 1 (MSN=2)     ----->    DATA MSG             ----->    PEER 2 (PMSN=2)
PEER 1 (MSN=2)     <-----    CONTROL MSG (ACK)    <-----    PEER 2 (PMSN=2)

PEER 1 (MSN=3)     ----->    DATA MSG             ----->    PEER 2 (PMSN=3)
PEER 1 (MSN=3)     <-----    CONTROL MSG (ACK)    <-----    PEER 2 (PMSN=3)

...

PEER 1 (MSN=255)   ----->   DATA MSG             ----->    PEER 2 (PMSN=255)
PEER 1 (MSN=255)   <-----   CONTROL MSG (ACK)    <-----    PEER 2 (PMSN=255)

PEER 1 (MSN=1)     ----->   DATA MSG             ----->    PEER 2 (PMSN=1)
PEER 1 (MSN=1)     <-----   CONTROL MSG (ACK)    <-----    PEER 2 (PMSN=1)
```

#### Corrupt message transmission with correction
A corrupt message, due to an invalid CRC checksum, is discarded and a resend is requested by replying with the Negative Acknowledge Control Message (NAK) to the communication peer. The MSN is incremented as always.

```
PEER 1 (MSN=2)     ----->    DATA MSG                 ----->    PEER 2 (PMSN=2)
PEER 1 (MSN=2)     <-----    ERROR CTRL MSG (NAK)     <-----    PEER 2 (PMSN=2)
PEER 1 (MSN=3)     ----->    DATA MSG                 ----->    PEER 2 (PMSN=3)
PEER 1 (MSN=3)     <-----    SUCCESS CTRL MSG (ACK)   <-----    PEER 2 (PMSN=3)
```

#### Message transmission with invalid PMSN
If the peer receives an Acknowledge Control message with an invalid not matching MSN as reply to a data message transmission, then this is handled as if a Negative Acknowledge Control Message was received.

```
PEER 1 (MSN=2)     ----->    DATA MSG                 ----->    PEER 2 (PMSN=10)
PEER 1 (MSN=2)     <-----    SUCCESS CTRL MSG (ACK)   <-----    PEER 2 (PMSN=10)  // Handle as if NAK is replied
PEER 1 (MSN=3)     ----->    DATA MSG                 ----->    PEER 2 (PMSN=3)
PEER 1 (MSN=3)     <-----    SUCCESS CTRL MSG (ACK)   <-----    PEER 2 (PMSN=3)
```

## 6. Append Data Flag
If the binary data is smaller than 1024 Bytes, then the complete data can be send within one message. The **append data flag** is set to false (**0x00**).

If the binary data is bigger than the maximum data body size of 1024 Bytes, then split the data body into maximum 1024 Byte data chunks and send them in multiple messages. The **append data flag** has to be set to true (**0x01**) to signalize the receiver, that the message has been split up into multiple messages.

### Sample - Sending binary data of size 879 bytes

```
|           SENDER            |

              ||
              ||
             \  /
              \/

|           MESSAGE           |
|-----------------------------|
| Append Data    = 0x00       |
| Data Body Size = 879 Bytes  |

              ||
              ||
             \  /
              \/

|          RECEIVER           |
```

### Sample - Sending binary data of size 2878 bytes

```
|           SENDER            |

              ||
              ||
             \  /
              \/

|           MESSAGE           |
|-----------------------------|
| Append Data    = 0x01       |
| Data Body Size = 1024 Bytes |

              ||
              ||
             \  /
              \/

|           MESSAGE           |
|-----------------------------|
| Append Data    = 0x01       |
| Data Body Size = 1024 Bytes |

              ||
              ||
             \  /
              \/

|           MESSAGE           |
|-----------------------------|
| Append Data    = 0x00       |
| Data Body Size = 830 Bytes  |

              ||
              ||
             \  /
              \/

|          RECEIVER           |
```

## 7. Error Handling
### 7.1 Invalid/Corrupted Data Messages
If a peer receives an invalid/corrupted data message, then the Negative Acknowledge Control Message is send to request a resend of the same message (Refer to the Control Message section for more information). This procedure is repeated until a valid message is received. Finally, on success, the Acknowledge Control Message is send.

```
PEER 1   ----->   DATA MESSAGE          ----->   PEER 2  (CRC checksum is invalid)
PEER 1   <-----   NAK CONTROL MESSAGE   <-----   PEER 2  (Request resend)
PEER 1   ----->   DATA MESSAGE          ----->   PEER 2  (CRC checksum is valid)
PEER 1   <-----   ACK CONTROL MESSAGE   <-----   PEER 2  (Continue with next message)
```

### 7.2 Invalid/Corrupted Control Messages
If a peer receives an invalid/corrupted control message, then the peer resends the data message just as if it received a Negative Acknowledge Control Message.

```
PEER 1   ----->   DATA MESSAGE              ----->   PEER 2
PEER 1   <-----   INVALID CONTROL MESSAGE   <-----   PEER 2
PEER 1   ----->   DATA MESSAGE              ----->   PEER 2  (Peer 1 resends the data message)
PEER 1   <-----   ACK CONTROL MESSAGE       <-----   PEER 2
```

### 7.3 Timeout - Read Data
Check the [8.2 Receive Data section](#82-receive-data) for more information.

### 7.4 Timeout - Control Messages
As soon as a peer has send a data message, it has to set a timeout of **5 seconds** and wait for a control message. If the control message was not received within the timeout, then the data message has to be resend, just as if the peer received a Negative Acknowledge Control Message.

```
PEER 1   ----->   DATA MESSAGE              ----->   PEER 2

[PEER 1 is waiting for a control message, but it will never receive one.]
[The timeout is reached and PEER 1 handles it as if it received a NAK control message]

PEER 1   ----->   DATA MESSAGE              ----->   PEER 2  (Peer 1 resends the data message)
PEER 1   <-----   ACK CONTROL MESSAGE       <-----   PEER 2
```

## 8. Message Transmission
### 8.1 Send Data
1. Construct the message body with the STX character as first character.
2. Increment the MSN and append it to the message body.
3. Split the binary data if it exceeds the maximum data length and set the append data flag depending on if the binary data has to be send in multiple messages (As defined in the Append Data Flag section).
4. Append the binary data body to the message body.
5. Calculate the CRC checksum of the message body and append it.
6. Append the ETX character to the message body.
7. Add the final message body to the send queue.
8. If there are any control messages available (control message queue), then send them first.
9. Now send the message body (from the send queue).
10. Wait for a control message (with a timeout as defined in the Error Handling section).
11. If an Acknowledge Control Message is received, then verify its checksum and the sequence number. The received MSN has to match with the MSN of the send data message. If this is a valid Acknowledge Control Message, then the data message transmission was successful.
12. Otherwise if a Negative Acknowledge Control Message is received or any other invalid control message, then resend the data message until an Acknowledge Control Message is received. (Handle the timeouts as defined in the Error Handling section)
13. Repeat this process if the binary data was split into multiple parts.

### 8.2 Receive Data
Data is read from e.g. a serial port within a loop. The received bytes are searched through for a STX, ACK or NAK control character, which indicates the start of a message block. Preceding bytes are dismissed. Bytes are read as long as the ETX end control character is found. If this process takes longer than **5 seconds**, then the received data is dismissed without sending control messages and the read process starts over again.

#### Receive Control Message
If a control message is received, then this control message has to be handled by the send data process. This can be solved by pushing it to a control message queue, which is read step-by-setp by the send processes.

#### Receive Data Message
As soon as a data message is received the following steps are processed:
1. Extract the CRC checksum and validate the message.
2. If the message is corrupted, then send a negative acknowledge control message and request a resend of the same message. Use the unknown message sequence number constant (UMSN) as message sequence number (MSN).
3. Extract the peer message sequence number (PMSN).
4. Extract the append data flag.
5. Extract the binary data body.
6. Push the binary data body to a temporary buffer.
7. Confirm a successful data transmission to the communication peer by sending the Acknowledge Control Message with the extracted PMSN.
8. If the append data flag signalizes, that the received binary data is not complete and is only a piece, then repeat these steps.
9. The final received binary data is now buffered in the temporary buffer.

## 9. Master/Slave Protocol
This asynchronous protocol can be easily transformed into a synchronous Master/Slave protocol.

The following additional rules apply:
1. The Slave does only send data messages as a response to a successful received data message from the Master.
2. The Master has to wait for a data message response from the Slave after a successful transmitted data message to the Slave.

**Important:** Multiple data messages to transmit bigger binary data chunks can be send to the Slave if the append data flag is set. The reply data message must be first send after a complete data transmission (multiple data messages received).

### 9.1 Samples
#### Successful data transmission

```
MASTER   ----->   DATA MESSAGE          ----->   SLAVE
MASTER   <-----   ACK CONTROL MESSAGE   <-----   SLAVE
MASTER   <-----   DATA MESSAGE          <-----   SLAVE
MASTER   ----->   ACK CONTROL MESSAGE   ----->   SLAVE
```

#### Successful data transmission with multiple data messages

```
MASTER   ----->   DATA MESSAGE (Append Data: 0x01)   ----->   SLAVE
MASTER   <-----   ACK CONTROL MESSAGE                <-----   SLAVE
MASTER   ----->   DATA MESSAGE (Append Data: 0x00)   ----->   SLAVE
MASTER   <-----   ACK CONTROL MESSAGE                <-----   SLAVE
MASTER   <-----   DATA MESSAGE (Append Data: 0x01)   <-----   SLAVE
MASTER   ----->   ACK CONTROL MESSAGE                ----->   SLAVE
MASTER   <-----   DATA MESSAGE (Append Data: 0x00)   <-----   SLAVE
MASTER   ----->   ACK CONTROL MESSAGE                ----->   SLAVE
```

#### Corrupted data transmission with correction #1

```
MASTER   ----->   DATA MESSAGE          ----->   SLAVE
MASTER   <-----   NAK CONTROL MESSAGE   <-----   SLAVE
MASTER   ----->   DATA MESSAGE          ----->   SLAVE
MASTER   <-----   ACK CONTROL MESSAGE   <-----   SLAVE
MASTER   <-----   DATA MESSAGE          <-----   SLAVE
MASTER   ----->   ACK CONTROL MESSAGE   ----->   SLAVE
```

#### Corrupted data transmission with correction #2

```
MASTER   ----->   DATA MESSAGE          ----->   SLAVE
MASTER   <-----   ACK CONTROL MESSAGE   <-----   SLAVE
MASTER   <-----   DATA MESSAGE          <-----   SLAVE
MASTER   ----->   NAK CONTROL MESSAGE   ----->   SLAVE
MASTER   <-----   DATA MESSAGE          <-----   SLAVE
MASTER   ----->   ACK CONTROL MESSAGE   ----->   SLAVE
```
