# Treadmill :running:

A mongodb backed persistent job runner for Go. 
Just a side project to help learn Go.
Get it? Treadmill? Job runner? .... moving on then.

Set it up:
Import it, create minimum options and create an instance of treadmill
```
import (
	"github.com/rob-johnston/treadmill"
	"github.com/rob-johnston/treadmill/options"
)

opts := options.Options{
		Database: "<Your DB name goes here>",
		Collection: "jobs",
		Client: client, // you need to provide a refernce to a mongodb client
}

tm := treadmill.NewTreadmill(opts)
 ```
 
 Then define yourself a job to run - eg a send email function.
 
 Your functions need to receive an empty interface.
 Within the function is where we will decode this empty interface in to a meaninful struct 
 you can actually use.
 
 ```
 sendEmail := func(data interface{}) error {
	
    // the data we require to run the job,
    // you provide this later when scheduling an individual job
type emailData struct {
    To string
    Subject string
    Body string
}
    
    // Use treadmills decode function to 
    // extract the data into your struct
    // You are responsible for making sure you
    // provide the right fields
		
    ed := &emailData{}
	err := tm.Decode(ed, data)
	if err != nil {
		log.Fatal(err)
	}
    
    // now our email data struct has been populated from whatever was in the empty interface
    // do email sending here...
    
return nil
}
```

  Now you can actually Define this task, name it whatever you like
  ```
	tm.Define("sendEmail", fetchData)
 ```
 
 Now you can run the treadmill with
 ```
 go tm.Run()
 ```
 
 And you can schedule jobs of that type by doing something like
 ```
emailData := struct{
    To string
    Subject string
    Body string
}{
    "yourcompany@recruiting.com"
    "Please hire me"
    "..."
} 
// this is the payload that will eventually be decoded into our emailData struct
// within our sendEmail function
  
  
 
err = tm.Schedule("2019-08-01 12:00:00", "fetchData", data)
 ```
