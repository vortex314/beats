		console.log("Javascript engine loaded "); 
		process = function(fields){
			fields.message = fields.syslog_message
			console.log(JSON.stringify(fields))
fields.attributes={}
fields.attributes.text = "just some Javascript code"
fields.metrics={}
fields.metrics."outside.temperature"=34.5
			fields.javascript="running in GO!"
//			var d = new Date(fields.timestamp)
//			console.log(" date " + d)
			return fields
		}
