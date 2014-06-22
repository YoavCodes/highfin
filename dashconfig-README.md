{
	"web": {
		"bins": [{
			"lang": "nodejs",
			"version": "0.10.28",
			"main": "./salmon.js",
			"watch": ["./"],
			"exclude": ["./salmon/node_modules", "./libs/node_modules", "./public/assets/css"],
			"npm": ["./salmon", "./libs"],
			
			/* path can be a full domain path or relative path. any request pointed at this container, either by sharkport or domain that matches path will be proxied by jellyfish to the specified port and hopefully to be handled by this binary. at our user's discretion" 
			
			Of course a binary can listen on multiple ports, and this setting is only used by jellyfish routing, it's within the bins block
			so it's easier to read, so client can remember that they've coded this binary to listen on that port.
			*/
			"endpoints": [{"path":"/0", "port": "8080"}], 

		}],
		"static": [{"path": "/", "dir": "./public"}]
		
		/* the domains block is for squid. It's in the appPart block so user can remember that this is a public container, which the given domains
		if pointed at squid via dns, will be proxied to this app's container.
		 */
		"domains": ["app1.test:80"]

		/*
		Instances is used for private networking between your apps.
		- if you set 5000, then one instance will be accessible from any app in a deploy via 127.0.0.1:5000
		note: the two instances may be physically deployed on separate machines in our infrastructure, but will still be able to tunnel to each-other
		via this mechanism.
		- if you set [5000, 5001], then you will have two separate instances each accessible directly via 127.0.0.1:5000 and 127.0.0.1:5001 respectively
		*/
		"instances": ["5000"],
		/*
		this block allows you to specify an ip address for round-robbining through instance ranges
		["0:1"] specifies a range of instance 0 to 1 in the zero-indexed array specified above in instances
		["2:5"] specifies index 2-5 ie: the third to the sixth entry
		too allow for real-time scaling of your app once in production
		[":"] specifies all instances
		["2:"] specifies instances 2 and up
		- if you set 0:1, 6000, then first two instances will be round-robbined anytime your app binary requests 127.0.0.1:6000
		- todo: consider allowing comma separated list and ranges. ie: 0,1,5:9,29.
		*/
		"balances": [{"6000": "0:1"}]
	}
}