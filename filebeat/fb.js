		console.log("Javascript engine loaded "); 
		process = function(fields){
			console.log(JSON.stringify(fields))
			fields.javascript="running in GO!"
			var d = new Date(fields.timestamp)
			console.log(" date " + d)
			return fields
		}
