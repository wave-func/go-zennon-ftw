package implementation

import (
	"crypto/sha256"
	"encoding/base64"
	"reflect"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	dcrdecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/ripemd160"

	"github.com/zenon-network/go-zenon/common"
	"github.com/zenon-network/go-zenon/common/types"
	"github.com/zenon-network/go-zenon/vm/constants"
)

var swapUtilsLog = common.EmbeddedLogger.New("contract", "swap-utils-log")

const (
	hashHeader          = "Zenon secp256k1 signature:"
	assetsMessage       = "ZNN swap retrieve assets"
	legacyPillarMessage = "ZNN swap retrieve legacy pillar"

	SwapRetrieveAssets       = 1
	SwapRetrieveLegacyPillar = 2
)

func toOldSignature(signature []byte) string {
	// transform signature in old znn-style signature
	header := signature[64]
	header += 31
	signature = append([]byte{header}, signature[0:64]...)
	return base64.StdEncoding.EncodeToString(signature)
}

func PubKeyToKeyId(pubKey []byte) []byte {
	// Parse the uncompressed public key and serialize as compressed
	parsedPubKey, err := secp256k1.ParsePubKey(pubKey)
	if err != nil {
		panic(err) 
	}
	compressed := parsedPubKey.SerializeCompressed()
	sha := sha256.New()
	sha.Write(compressed)
	ripe := ripemd160.New()
	ripe.Write(sha.Sum(nil))
	return ripe.Sum(nil)
}

func PubKeyToKeyIdHash(pubKey []byte) types.Hash {
	keyId := PubKeyToKeyId(pubKey)
	sha := sha256.New()
	sha.Write(keyId)
	return types.BytesToHashPanic(sha.Sum(nil))
}

// SignRetrieveAssetsMessage is used for in contract tests
func SignRetrieveAssetsMessage(address types.Address, prv []byte, pub string) (string, error) {
	// config message & verify against expected message
	message := GetSwapMessage(assetsMessage, pub, address)

	// sign message
	privKey := secp256k1.PrivKeyFromBytes(prv)
	sig := dcrdecdsa.SignCompact(privKey, message, false)
	
	// Convert from compact [V || R || S] to ethereum format [R || S || V]
	signature := make([]byte, 65)
	copy(signature[0:32], sig[1:33])   // R
	copy(signature[32:64], sig[33:65]) // S
	signature[64] = sig[0] - 27        // V (adjust back)
	
	return toOldSignature(signature), nil
}

// SignLegacyPillarMessage is used for in contract tests
func SignLegacyPillarMessage(address types.Address, prv []byte, pub string) (string, error) {
	// config message & verify against expected message
	message := GetSwapMessage(legacyPillarMessage, pub, address)

	// sign message
	privKey := secp256k1.PrivKeyFromBytes(prv)
	sig := dcrdecdsa.SignCompact(privKey, message, false)
	
	// Convert from compact [V || R || S] to ethereum format [R || S || V]
	signature := make([]byte, 65)
	copy(signature[0:32], sig[1:33])   // R
	copy(signature[32:64], sig[33:65]) // S
	signature[64] = sig[0] - 27        // V (adjust back)
	
	return toOldSignature(signature), nil
}

func serializeString(txt string) []byte {
	y := append([]byte(""), byte(len(txt)))
	return append(y, []byte(txt)...)
}

func GetSwapMessage(operationMessage string, pubKey string, addr types.Address) []byte {
	var data []byte
	data = append(data, serializeString(hashHeader)...)
	data = append(data, serializeString(operationMessage+" "+pubKey+" "+addr.String())...)
	a := sha256.Sum256(data)
	b := sha256.Sum256(a[:])
	return b[:]
}

func CheckSwapSignature(messageType int, addr types.Address, pubKeyStr string, signatureStr string) (bool, error) {
	pubKey, err := base64.StdEncoding.DecodeString(pubKeyStr)
	if err != nil {
		swapUtilsLog.Debug("swap-utils-error", "reason", "malformed-pubKey")
		return false, constants.ErrInvalidB64Decode
	}
	if len(pubKey) != 65 {
		swapUtilsLog.Debug("swap-utils-error", "reason", "invalid-pubKey-length")
		return false, constants.ErrInvalidB64Decode
	}

	sig, err := base64.StdEncoding.DecodeString(signatureStr)
	if err != nil {
		swapUtilsLog.Debug("swap-utils-error", "reason", "malformed-signature")
		return false, constants.ErrInvalidB64Decode
	}
	if len(sig) != 65 {
		swapUtilsLog.Debug("swap-utils-error", "reason", "invalid-signature-length")
		return false, constants.ErrInvalidSignature
	}

	var operationMessage string
	if messageType == SwapRetrieveAssets {
		operationMessage = assetsMessage
	} else if messageType == SwapRetrieveLegacyPillar {
		operationMessage = legacyPillarMessage
	} else {
		swapUtilsLog.Debug("swap-utils-error", "reason", "invalid-operation")
		return false, constants.ErrInvalidSwapCode
	}

	message := GetSwapMessage(operationMessage, pubKeyStr, addr)
	swapUtilsLog.Debug("swap-utils-log", "expected-message", hexutil.Encode(message))

	// Transform signature from Old Znn-style to go secp256k1 signature
	header := sig[0]
	header -= 31
	sig = append(sig, header)
	sig = sig[1:]

	// Convert signature from [R || S || V] to [V || R || S] format for RecoverCompact
	compactSig := make([]byte, 65)
	compactSig[0] = sig[64] + 27 // V at front, add 27 for compact format
	copy(compactSig[1:33], sig[0:32])   // R
	copy(compactSig[33:65], sig[32:64]) // S

	recoveredKey, _, err := dcrdecdsa.RecoverCompact(compactSig, message)
	if err != nil {
		swapUtilsLog.Debug("swap-utils-error", "reason", err)
		return false, constants.ErrInvalidSignature
	}
	recoveredPubKey := recoveredKey.SerializeUncompressed()
	if !reflect.DeepEqual(pubKey, recoveredPubKey) {
		swapUtilsLog.Debug("swap-utils-error", "reason", "invalid-signature")
		return false, constants.ErrInvalidSignature
	}

	return true, nil
}
