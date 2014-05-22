Vagrant.configure("2") do |config|
 
  config.vm.box = "bytepark/debian-7.4"
  config.vm.provision :shell, :path => "vagrant_setup.sh"


  config.vm.define "squid" do |squid|
  	squid.vm.network :private_network, ip: "10.10.10.5"
  	squid.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "160"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  	
  end

  config.vm.define "octopus" do |octopus|
  	octopus.vm.network :private_network, ip: "10.10.10.50"
  	octopus.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "160"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  end

  config.vm.define "shark" do |shark|
  	shark.vm.network :private_network, ip: "10.10.10.10"
  	shark.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "160"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  end

  config.vm.define "guppy" do |guppy|
  	guppy.vm.network :private_network, ip: "10.10.10.100"
  	guppy.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "160"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  end

  

end
