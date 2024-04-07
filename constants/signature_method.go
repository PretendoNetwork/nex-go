package constants

// SignatureMethod is an implementation of the nn::nex::PRUDPMessageInterface::SignatureMethod enum.
//
// The signature method is used as part of the packet signature calculation process. It determines
// what data is used and from where when calculating the packets signature.
//
// Currently unused. Implemented for future use and dodumentation/note taking purposes.
//
// The following notes are derived from Xenoblade on the Wii U. Many details are unknown.
//
// Based on the `nn::nex::PRUDPMessageV1::CalcSignatureHelper` (`__CPR210__CalcSignatureHelper__Q3_2nn3nex14PRUDPMessageV1FPCQ3_2nn3nex6PacketQ4_2nn3nex21PRUDPMessageInterface15SignatureMethodPCQ3_2nn3nex3KeyQJ68J3nex6Stream4TypePCQ3_2nn3nex14SignatureBytesRQ3_2nn3nexJ167J`)
// function:
//
// There appears to be 9 signature methods. Methods 0, 2, 3, and 9 seem to do nothing.  Method 1
// seems to calculate the signature using the connection address. Methods 4-8 calculate the signature
// using parts of the packet.
//
//   - Method 0: Calls `func_0x04b10f90` and bails immediately?
//   - Method 1: Seems to calculate the signature using ONLY the connections address? It uses the values
//     from `nn::nex::InetAddress::GetAddress` and `nn::nex::InetAddress::GetPortNumber`, among others.
//     It does NOT follow the same code path as methods 4-9
//   - Method 2: Unknown. Bails without doing anything
//   - Method 3: Unknown. Bails without doing anything
//
// Methods 4-8 build the signature from one or many parts of the packet
//
//   - Methods 4-8: Use the value from `nn::nex::Packet::GetHeaderForSignatureCalc`?
//   - Methods 5-8: Use whatever is passed as `signature_bytes_1`, but only if:
//     1. `signature_bytes_1` is not empty.
//     2. The packet type is not `SYN`.
//     3. The packet type is not `CONNECT`.
//     4. The packet type is not `USER` (?).
//     5. `type_flags & 0x200 == 0`.
//     6. `type_flags & 0x400 == 0`.
//   - Method 6: Use an optional "key", if not null
//   - If method 7 is used, 2 local variables are set to 0. Otherwise they get set the content pointer
//     and size of the calculated signature buffer. In both cases another local variable is set to
//     `packet->field_0x94`, and then some checks are done on it before it's set to the packets payload?
//   - Method 8: 16 random numbers generated and appended to `signature_bytes_2`
//   - Method 9: The signature seems ignored entirely?
type SignatureMethod uint8

const (
	// SignatureMethod0 is an unknown signature type
	SignatureMethod0 SignatureMethod = iota

	// SignatureMethodConnectionAddress seems to indicate the signature is based on the connection address
	SignatureMethodConnectionAddress

	// SignatureMethod2 is an unknown signature type
	SignatureMethod2

	// SignatureMethod3 is an unknown signature type
	SignatureMethod3

	// SignatureMethod4 is an unknown signature method
	SignatureMethod4

	// SignatureMethod5 is an unknown signature method
	SignatureMethod5

	// SignatureMethodUseKey seems to indicate the signature uses the provided key value, if not null
	SignatureMethodUseKey

	// SignatureMethod7 is an unknown signature method
	SignatureMethod7

	// SignatureMethodUseEntropy seems to indicate the signature includes 16 random bytes
	SignatureMethodUseEntropy

	// SignatureMethodIgnore seems to indicate the signature is ignored
	SignatureMethodIgnore
)
