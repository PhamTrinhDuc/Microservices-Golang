User đăng nhập: username + password => Auth service => ký jwt gửi lại cho user.

User gửi lại data + JWT cho server. Server dùng public key để validate data + JWT này.

1. Nếu đúng chuẩn như Auth service đã gửi lại cho user(không bị sửa đổi) => pass
2. Nếu bị sửa 1 trường như role: admin => khác với Auth service từng ký => chặn

Lưu ý: JWT chống sửa đổi dữ liệu không chống hacker vào account, nhưng nếu hacker lấy được JWT và gửi data + JWT đúng như vậy thì sẽ vào được account của user
---

Hãy nhìn vào sự khác biệt giữa việc **Đọc** và **Ghi**:

### 1. Biết "Chữ ký thật" không giúp bạn "Ký giả"
Việc hacker có Public Key và dùng nó để validate một Token cũng giống như việc bạn cầm một tờ hóa đơn có chữ ký của Giám đốc, sau đó bạn lên mạng tìm mẫu chữ ký công khai của ông ấy để đối chiếu.
* **Kết quả đối chiếu:** Bạn xác nhận "Ừ, đây đúng là chữ ký của Giám đốc".
* **Hành động tiếp theo:** Bạn muốn sửa số tiền trên hóa đơn từ 1 triệu thành 100 triệu. 
* **Vấn đề:** Khi bạn sửa con số, cái "chữ ký" cũ không còn khớp với nội dung mới nữa. Để hóa đơn 100 triệu đó có hiệu lực, bạn cần **ký lại**. Nhưng bạn không có tay của Giám đốc (Private Key), bạn chỉ có cái ảnh chụp chữ ký (Public Key) để nhìn thôi. 



### 2. Public Key là "Ống nhòm", không phải "Cây bút"
* **Validate (Xác minh):** Là hành động thụ động. Bạn chỉ kiểm tra xem dữ liệu có bị sửa đổi hay không. Hacker dùng Public Key để validate token của chính hắn thì cũng chỉ nhận được kết quả: "Token này hợp lệ". Vậy thì sao? Hắn vốn đã biết nó hợp lệ rồi.
* **Forge (Giả mạo):** Là hành động chủ động. Hacker muốn tạo ra một Token mới với `role: admin`. Để làm việc này, hắn cần thuật toán mã hóa chạy trên **Private Key**. Public Key hoàn toàn bất lực, không thể dùng để tạo ra chữ ký mới.

---

### Tóm lại, đây là kịch bản của Hacker:

1.  **Hacker bắt được 1 JWT của User A:** Hắn dùng Public Key để validate. Server báo: "Hợp lệ, đây là User A". Hacker vẫn chỉ là User A.
2.  **Hacker sửa nội dung JWT thành User B:** Hắn dùng Public Key để validate lại. Thuật toán báo: **"SAI! Chữ ký không khớp với nội dung"**.
3.  **Hacker muốn ký lại cho khớp:** Hắn cần Private Key. Nhưng hắn không có. Hắn bị kẹt ở đây.

**Cái "Bảo mật" ở đây không phải là giấu Token, mà là khiến Token trở nên "Bất khả xâm phạm về nội dung".**

Bạn đang lo lắng rằng nếu hacker biết cách validate (biết Public Key), hắn sẽ tìm ra kẽ hở để "ngược dòng" tìm ra Private Key đúng không? Trong toán học hiện đại, việc này mất khoảng... vài tỷ năm với máy tính hiện nay.

Vậy, nếu tôi nói rằng thực tế người ta còn **công khai** luôn cả Public Key trên một đường link URL cho cả thế giới tải về (gọi là JWKS), bạn có thấy vô lý không? Hay bạn bắt đầu thấy tin vào độ mạnh của toán học rồi?