package idgen

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// GenerateTimeRandomID สร้าง ID แบบ int64
func GenerateTimeRandomID() int64 {
	timestamp := time.Now().UnixNano() >> 32
	randomPart := rand.Int63() & 0xFFFFFFFF
	return (timestamp << 32) | randomPart
}

// GenerateTimeID สร้าง ID แบบ int (ใช้ timestamp เป็นหลัก)
func GenerateTimeID() int {
	// ใช้ timestamp Unix วินาที (int64) แปลงเป็น int (int32/64 ขึ้นกับระบบ)
	return int(time.Now().Unix())
}

// GenerateTimeRandomIDBase36 คืนค่า ID เป็น base36 string
func GenerateTimeRandomIDBase36() string {
	id := GenerateTimeRandomID()
	return strconv.FormatInt(id, 36) // แปลงเลขฐาน 10 -> 36
}

// GenerateUUIDLikeID คืนค่าเป็น string แบบ UUID-like (แต่ไม่ใช่ UUID จริง)
func GenerateUUIDLikeID() string {
	id := GenerateTimeRandomID()

	// แปลง int64 เป็น hex string ยาว 16 ตัว (64 bit)
	hex := fmt.Sprintf("%016x", uint64(id))

	// สร้าง UUID-like string รูปแบบ 8-4-4-4-12
	// แต่มีแค่ 16 hex chars แบ่งคร่าวๆ: 8-4-4 (เหลือไม่พอจริงๆ)
	// ดังนั้นเราจะเติม random เพิ่มเพื่อครบ 32 hex (128 bit) เหมือน UUID

	randPart := fmt.Sprintf("%016x", rand.Uint64())

	uuidLike := strings.Join([]string{
		hex[0:8],
		hex[8:12],
		hex[12:16],
		randPart[0:4],
		randPart[4:16],
	}, "-")

	return uuidLike
}

// ก่อน Go 1.20	ต้องเรียก เพื่อให้ได้เลขสุ่มไม่ซ้ำ
// func init() {
// 	rand.Seed(time.Now().UnixNano())
// }
