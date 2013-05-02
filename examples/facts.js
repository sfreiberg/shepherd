// Example of currently available facts
console.log(facts.hostname)
console.log(facts.domain)
console.log(facts.fqdn)
console.log(facts.cpus)
console.log(facts.os)
for (interface in facts.interfaces) {
	console.log("##########################")
	console.log(interface)
	console.log(facts.Interfaces[interface]["mtu"])
	console.log(facts.Interfaces[interface]["hardware_addr"])
	var addresses = facts.Interfaces[interface]["addresses"]
	for (address in addresses) {
		console.log(addresses[address]["address"])
	}
	
}
