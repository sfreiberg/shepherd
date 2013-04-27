// Example of currently available facts
console.log(facts.Hostname)
console.log(facts.Domain)
console.log(facts.Fqdn)
console.log(facts.Cpus)
console.log(facts.Os)
for (interface in facts.Interfaces) {
	console.log("##########################")
	console.log(interface)
	console.log(facts.Interfaces[interface]["MTU"])
	console.log(facts.Interfaces[interface]["HardwareAddr"])
	var addresses = facts.Interfaces[interface]["Addresses"]
	for (address in addresses) {
		console.log(addresses[address]["Address"])
	}
	
}
