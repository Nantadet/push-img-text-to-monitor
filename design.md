# Instagram Live Display System - Design

## 1. สรุปแนวคิด

ระบบนี้เป็น Live Display Platform ที่ให้ Guest ส่งลิงก์ Instagram พร้อมข้อความสั้น ๆ เข้ามาในระบบ จากนั้นระบบจะดึงรูปจากลิงก์ IG มาแสดงบนจอหลักพร้อมข้อความและลิงก์วาร์ปไปยัง IG นั้น

Admin จะเป็นคนควบคุมจอแสดงผลผ่านหน้า dashboard โดยดูข้อมูลที่ user ส่งเข้ามาทั้งหมด และสั่งข้ามรายการ เพิ่มเวลา หรือลดเวลาให้รายการที่กำลังแสดงได้แบบ realtime



> หมายเหตุ: การดึงรูปจาก Instagram ควรทำผ่านวิธีที่ถูกต้อง เช่น Instagram Graph API, oEmbed, หรือ metadata ที่ระบบเข้าถึงได้อย่างถูกสิทธิ์ ไม่ควรออกแบบให้ scrape ข้อมูลที่ละเมิดเงื่อนไขของแพลตฟอร์ม

## 2. ผู้ใช้งาน

| Role | หน้าที่หลัก | ต้อง Login |
| --- | --- | --- |
| Guest | ส่งลิงก์ IG และข้อความ | ไม่ต้อง |
| Admin | ควบคุมคิวและหน้าจอแสดงผล | ต้อง |

### Guest ทำอะไรได้

- วางลิงก์ Instagram
- พิมพ์ข้อความประกอบ
- Preview รูปที่ระบบดึงจาก IG
- ส่งข้อมูลเข้าคิวแสดงผล
- ได้ลิงก์วาร์ป IG ติดไปกับรายการที่ส่ง

### Admin ทำอะไรได้

- Login เข้าหน้า Admin
- ดูข้อมูลทั้งหมดที่ user ส่งเข้ามา
- ดูรายการที่กำลังแสดงอยู่
- ข้ามรูปและข้อความที่กำลังแสดง
- เพิ่มเวลาแสดงผลให้รายการปัจจุบัน
- ลดเวลาแสดงผลของรายการปัจจุบัน
- เลือกรายการใดรายการหนึ่งให้ขึ้นจอ
- ลบรายการที่ไม่เหมาะสม
- ดูสถานะ realtime ของคิว

## 3. ประโยคสรุปแอป

Guest ส่งลิงก์ Instagram และข้อความเข้ามา ส่วนระบบดึงรูปจาก IG มาแสดงบนจอแบบ realtime พร้อมให้ Admin ดูข้อมูลทั้งหมดใน dashboard และควบคุมคิวด้วยการเพิ่มเวลา ลดเวลา หรือข้ามรายการได้

## 4. Flow หลัก

### Guest Flow

```txt
เข้าเว็บ
  -> กด Guest
  -> วางลิงก์ Instagram
  -> ระบบตรวจสอบลิงก์
  -> Backend ดึงรูป/metadata จาก IG
  -> Guest พิมพ์ข้อความ
  -> Preview รูป + ข้อความ + วาร์ป IG
  -> กดส่ง
  -> Backend บันทึกข้อมูลลง MongoDB
  -> Broadcast รายการใหม่ผ่าน WebSocket
  -> รายการเข้า Queue รอ Admin หรือ Auto Display
```

### Admin Flow

```txt
เข้าเว็บ
  -> กด Admin
  -> Login
  -> เข้า Dashboard
  -> ดูข้อมูลและ Queue รายการ IG ที่ user ส่งมา
  -> เลือกรายการให้แสดง / ข้ามรายการ / เพิ่มเวลา / ลดเวลา / ลบรายการ
  -> Display Screen อัปเดตแบบ realtime
```

### Display Flow

```txt
มีรายการใน Queue
  -> แสดงรูปจาก IG
  -> แสดงข้อความจาก Guest
  -> แสดงปุ่มหรือ QR วาร์ป IG
  -> Frontend นับเวลาถอยหลังเองจาก displayedAt + displayMinutes
  -> ถ้าหมดเวลา ไปยังรายการถัดไป
  -> ถ้า Admin กด Skip ให้เปลี่ยนทันที
  -> ถ้า Admin กด Add Time ให้เพิ่มเวลาแสดงผล
  -> ถ้า Admin กด Reduce Time ให้ลดเวลาแสดงผล
```

## 5. หน้าที่ต้องมี

| Page | ผู้ใช้ | หน้าที่ |
| --- | --- | --- |
| Guest Submit Page | Guest | ส่งลิงก์ IG และข้อความ |
| Admin Login Page | Admin | Login เข้าระบบ |
| Admin Dashboard | Admin | ดูข้อมูลที่ user ส่งเข้ามาและควบคุม Queue กับ Display |
| Display Screen | จอหลัก | แสดงรูป IG, ข้อความ และวาร์ป IG |


### Guest Submit Page

- Instagram link input
- Message input
- Preview รูปจาก IG
- Preview ข้อความ
- Preview วาร์ป IG
- Submit button
- Error state ถ้าลิงก์ IG ไม่ถูกต้องหรือดึงรูปไม่ได้

### Admin Login Page

- Username input
- Password input
- Login button
- Error state เมื่อ login ไม่ผ่าน

### Admin Dashboard

- Current Display panel
- User submission list
- Queue list
- Preview รูปและข้อความของแต่ละรายการ
- ปุ่ม Show Now
- ปุ่ม Skip Current
- ปุ่ม Add Time
- ปุ่ม Reduce Time
- ปุ่ม Delete
- เวลาคงเหลือของรายการที่กำลังแสดง
- สถานะ WebSocket connection
- Dashboard เรียงรายการ queued ตาม `createdAt` เก่าสุดขึ้นก่อน

### Display Screen

- แสดงรูปจาก Instagram ขนาดใหญ่
- แสดงข้อความของ Guest
- แสดง Instagram username หรือ URL
- แสดง QR code หรือปุ่มวาร์ป IG
- Countdown timer
- Live update เมื่อ Admin เปลี่ยนรายการ

## 6. ข้อมูลหลักในระบบ

### Collection: display_items

```json
{
  "_id": "ObjectId",
  "igUrl": "string",
  "igImageUrl": "string",
  "igUsername": "string",
  "message": "string",
  "status": "queued | displaying | skipped | displayed",
  "createdAt": "date",
  "displayedAt": "date | null",
  "finishedAt": "date | null"
}
```

| Field | Type | รายละเอียด |
| --- | --- | --- |
| `_id` | ObjectId | ใช้ระบุรายการ |
| `igUrl` | string | ลิงก์ Instagram ที่ Guest ส่งมา |
| `igImageUrl` | string | URL รูปที่ดึงจาก IG |
| `igUsername` | string | ชื่อบัญชี IG ถ้าดึงได้ |
| `message` | string | ข้อความจาก Guest |
| `status` | string | สถานะของรายการใน Queue |
| `displayMinutes` | number | เวลาแสดงผลรวมของรายการในหน่วยนาที |
| `createdAt` | date | เวลาที่ Guest ส่งรายการ |
| `displayedAt` | date | เวลาที่เริ่มแสดง |
| `finishedAt` | date | เวลาที่รายการจบการแสดงหรือถูกข้าม |

### Timer Logic

- Frontend เป็นคนคำนวณเวลาถอยหลังเอง
- Database ไม่ต้อง update ค่า countdown ทุกวินาที
- Frontend คำนวณจาก `displayMinutes` และ `displayedAt`
- ถ้า Admin เพิ่มหรือลดเวลา ระบบ update ค่า `displayMinutes` แล้ว broadcast ค่าใหม่ให้ frontend คำนวณต่อ
- ถ้า status ไม่ใช่ `displaying` frontend ไม่ต้องนับต่อ

## 7. สิทธิ์การใช้งาน

| Feature | Guest | Admin |
| --- | --- | --- |
| ส่งลิงก์ IG | ได้ | ไม่ได้ |
| พิมพ์ข้อความ | ได้ | ไม่ได้ |
| Preview รูป IG | ได้ | ได้ |
| ดู Queue ทั้งหมด | ไม่ได้ | ได้ |
| เลือกรายการให้แสดง | ไม่ได้ | ได้ |
| ข้ามรายการปัจจุบัน | ไม่ได้ | ได้ |
| เพิ่มเวลาแสดงผล | ไม่ได้ | ได้ |
| ลบรายการ | ไม่ได้ | ได้ |
| ดู Display Screen | ได้ | ได้ |

## 8. ระบบ Queue และเวลาแสดงผล

### กติกา Queue

- รายการใหม่จะถูกเพิ่มเข้า Queue ด้วยสถานะ `queued`
- ถ้าไม่มีรายการกำลังแสดง ระบบจะเลือกรายการที่มี `status=queued` และ `createdAt` เก่าสุดขึ้นแสดงอัตโนมัติ
- รายการที่กำลังแสดงจะมีสถานะ `displaying`
- เมื่อหมดเวลา รายการจะเปลี่ยนเป็น `displayed` และบันทึก `finishedAt`
- ถ้า Admin กด Skip รายการจะเปลี่ยนเป็น `skipped` และบันทึก `finishedAt`
- หลังจากรายการปัจจุบันจบ ระบบจะเลือกตัวถัดไปจากรายการ `status=queued` ที่ `createdAt` เก่าสุด

### Timestamp Rules

- `createdAt` คือเวลาที่ user ส่งข้อมูลเข้าระบบ
- `displayedAt` คือเวลาที่รายการถูกนำขึ้นจอจริง
- `finishedAt` คือเวลาที่รายการจบ, ถูกข้าม, หรือออกจากจอ
- Queue order อิงจาก `createdAt` ของรายการที่ยังมี `status=queued`

### การเพิ่มเวลา

Admin สามารถเพิ่มเวลาให้รายการที่กำลังแสดง เช่น:

- Add 1 minute
- Add 3 minutes
- Add 5 minutes

เมื่อเพิ่มเวลา ระบบจะ update `displayMinutes` และ broadcast ไปยัง Display Screen ทันที โดย frontend จะคำนวณ countdown ต่อจากค่าใหม่

### การลดเวลา

Admin สามารถลดเวลาให้รายการที่กำลังแสดง เช่น:

- Reduce 1 minute
- Reduce 3 minutes

เมื่อเวลาถูกลด ระบบจะ update `displayMinutes` ทันที และถ้า frontend คำนวณแล้วเวลาคงเหลือเหลือ 0 หรือน้อยกว่า 0 ให้ frontend เรียก backend เพื่อเปลี่ยน status ของรายการปัจจุบันและเลือกรายการ queued ที่เก่าสุดขึ้นจออัตโนมัติ

### การข้ามรายการ

เมื่อ Admin กด Skip:

1. รายการปัจจุบันเปลี่ยนเป็น `skipped`
2. ระบบเลือกรายการ `status=queued` ที่ `createdAt` เก่าสุด
3. Display Screen เปลี่ยนรูปและข้อความทันที
4. Broadcast event ไปยังทุก client ที่เชื่อมต่ออยู่

## 9. API Design

### Guest APIs

```http
POST /items/preview
```

รับ `igUrl` แล้วดึงข้อมูล preview จาก Instagram เช่นรูป, username, title หรือ metadata ที่จำเป็น

```http
POST /items
```

สร้างรายการใหม่จาก `igUrl`, `igImageUrl`, `igUsername`, และ `message`

### Admin APIs

```http
GET /admin/items
GET /admin/items/current
POST /admin/items/:id/show
POST /admin/items/current/skip
POST /admin/items/current/add-time
POST /admin/items/current/reduce-time
POST /admin/items/current/finish
DELETE /admin/items/:id
POST /admin/login
```

### Display APIs

```http
GET /display/current
GET /display/queue
```

`GET /display/queue` ควรเรียง `status=queued` ตาม `createdAt ASC`

### Realtime

```txt
WebSocket route: /ws
```

Events:

```txt
item-created
item-showing
item-skipped
item-time-added
item-time-reduced
item-deleted
queue-updated
display-ended
```

## 10. Instagram Link Processing

### Input

Guest ส่งลิงก์ Instagram เช่น:

```txt
https://www.instagram.com/p/xxxx/
https://www.instagram.com/reel/xxxx/
```

### Backend Process

```txt
Receive IG URL
  -> Validate URL format
  -> Fetch metadata with approved method
  -> Extract image URL
  -> Extract username if available
  -> Return preview to frontend
```

### Error Cases

- ลิงก์ไม่ใช่ Instagram
- ลิงก์เป็น private post
- ดึงรูปไม่ได้
- Instagram rate limit
- รูปหมดอายุหรือ URL ใช้งานไม่ได้

### Fallback

ถ้าดึงรูปไม่ได้ ระบบควร:

- แจ้ง Guest ว่าไม่สามารถโหลดรูปได้
- ให้เปลี่ยนลิงก์ใหม่
- ไม่สร้างรายการเข้า Queue จนกว่าจะ preview สำเร็จ

## 11. MVP

### ต้องมี

#### Backend

- Fiber server
- MongoDB connection
- Auth สำหรับ Admin login
- API preview ลิงก์ IG
- API สร้าง display item
- Queue management
- Skip current item
- Add display time
- Reduce display time
- Move to oldest queued item when current item ends
- Delete item
- WebSocket realtime

#### Frontend

- Home page
- Guest submit page
- IG link preview
- Message input
- Admin login page
- Admin dashboard
- Display screen

#### System

- แสดงรูปจาก IG
- แสดงข้อความ Guest
- วาร์ปไป IG ได้
- Queue realtime
- Countdown timer
- Frontend local countdown
- Admin skip
- Admin add time
- Admin reduce time

### เพิ่มทีหลังได้

- QR code สำหรับวาร์ป IG
- Auto moderation
- Animation ตอนเปลี่ยนรูป
- Sound effect
- Online user count
- Multiple display screens
- Analytics
- Pin item
- Shuffle queue

## 12. เป้าหมายสุดท้าย

สร้างระบบ Live Instagram Display ที่:

- ให้ Guest ส่งลิงก์ IG และข้อความได้ง่าย
- ดึงรูปจาก IG มาแสดงบนจอหลัก
- แสดงข้อความและวาร์ป IG พร้อมกัน
- ให้ Admin คุมจอได้แบบ realtime
- ข้ามรายการที่ไม่ต้องการได้ทันที
- เพิ่มเวลาให้รายการที่ต้องการโชว์นานขึ้นได้
- ลดเวลาให้รายการที่ไม่ต้องการแสดงนานได้
- รองรับการต่อยอดเป็นจอ event หรือ interactive screen ได้

ส่ง tiktok reel ได้เพื่อดึงทั้งคริปและเสียงเลย แล้วก็music Lik url ได้
