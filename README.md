# minitel in Go


```go

payload := minitel.Notification{
	Title: "Your DB is on fire!",
	Body: "...",
}

payload.Target.ID = "84838298-989d-4409-b148-6abef06df43f"
payload.Target.Type = minitel.App

payload.Action.Label = "View Invoice"
payload.Action.URL = "https://view.your.invoice/yolo"

client, err := minitel.New("https://user:pass@telex.heroku.com")
if err != nil {
	log.Fatal(err)
}
client.Notify(payload)
```
