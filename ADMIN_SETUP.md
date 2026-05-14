# วิธีสร้าง Admin Account

หน้า **Admin** (`/admin`) เปิดให้ **Login** อย่างเดียว ไม่มีปุ่มสมัครบนหน้าเว็บแล้ว  
ถ้าต้องการสร้างบัญชี Admin ให้ยิง API ผ่าน **Postman** หรือเครื่องมืออื่น ๆ ตามขั้นตอนด้านล่าง

---

## Endpoint

```
POST /auth/register
```

Base URL ขึ้นอยู่กับ environment:
- Local: `http://localhost:3000`
- Production: URL ที่ deploy ไว้

---

## Headers

| Header | Value |
|--------|-------|
| `Content-Type` | `application/json` |

---

## Request Body (JSON)

```json
{
  "username": "admin",
  "password": "password123"
}
```

### Validation Rules

| Field | กฎ |
|-------|-----|
| `username` | **required**, ความยาว 3-60 ตัวอักษร |
| `password` | **required**, ความยาว 6-128 ตัวอักษร |

---

## ขั้นตอนใน Postman

1. เปิด Postman → สร้าง **New Request**
2. เลือก Method เป็น **POST**
3. ใส่ URL: `http://localhost:3000/auth/register`
4. ไปที่แท็บ **Headers** → เพิ่ม:
   - `Content-Type` = `application/json`
5. ไปที่แท็บ **Body** → เลือก **raw** → เลือก **JSON**  → ใส่:
   ```json
   {
     "username": "your_admin_name",
     "password": "your_secure_password"
   }
   ```
6. กด **Send**

---

## Response

### ✅ สำเร็จ — `201 Created`

```json
{
  "id": "68249abc...",
  "username": "admin"
}
```

> สมัครเสร็จแล้ว ไปที่หน้า `/admin` แล้ว login ด้วย username/password ที่เพิ่งสร้างได้เลย

### ❌ บัญชีซ้ำ — `409 Conflict`

```json
{
  "error": "user already exists"
}
```

> username นี้มีในระบบแล้ว ให้เปลี่ยน username แล้วยิงใหม่

### ❌ ข้อมูลไม่ถูกต้อง — `400 Bad Request`

```json
{
  "error": "Key: 'RegisterDTO.Username' Error:Field validation for 'Username' failed on the 'required' tag"
}
```

> ตรวจสอบว่าใส่ `username` และ `password` ครบ และความยาวถูกต้องตามกฎด้านบน

---

## หมายเหตุ

- ระบบไม่มี rate limiting บน auth endpoints แต่ควรสร้าง admin แค่ครั้งเดียวพอ
- หลังจากสร้าง admin แล้ว ให้ลบหรือเก็บ credentials ไว้อย่างปลอดภัย
- ถ้าลืม password ไม่มีระบบ reset ต้องลบ user จาก database แล้วสร้างใหม่
