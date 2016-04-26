# minitel in Go


```go

payload := minitel.Payload{
	Title: "Your DB is on fire!",
	Body: "...",
}

payload.Target.Id = "84838298-989d-4409-b148-6abef06df43f"
payload.Target.Type = minitel.App

payload.Action.Label = "View Invoice"
payload.Action.URL = "https://view.your.invoice/yolo"

client := minitel.New("https://user:pass@telex.heroku.com")
client.Notify(payload)
```
