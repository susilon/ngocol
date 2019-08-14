# ngocol
NgoCol a.k.a Ngobrol di COnsoLe, Proof of Concept gRPC programming dengan Go,

  - Type some Markdown on the left
  - See HTML in the right
  - Magic

## Server
```sh
$ ng_server [optional:port]
```
Contoh : `ng_server 5110`
Tanpa parameter port maka akan menggunakan default port : 50051

## Client
```sh
$ ngocol
```
Pertama kali akan menanyakan :
- Server address:port
- Nickname
- Status


untuk mengganti config setelah running : /c

untuk melihat list user yang online : /l

untuk keluar program : /q
