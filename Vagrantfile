Vagrant.configure("2") do |config|
 
  config.vm.box = "bytepark/debian-7.4"
  config.vm.provision "shell", path: "vagrant_setup.sh"
  
  config.vm.provider :virtualbox do |vb|
      vb.customize ["modifyvm", :id, "--memory", "160"]
      vb.customize ["modifyvm", :id, "--vram", "1"]
      vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
  end



  config.vm.define "squid" do |conf|
  	conf.vm.network :private_network, ip: "10.10.10.5"
  end

  config.vm.define "octopus" do |conf|
  	conf.vm.network :private_network, ip: "10.10.10.50"
  end

  config.vm.define "shark0" do |conf|
  	conf.vm.network :private_network, ip: "10.10.10.10"
    conf.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "256"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  end
  config.vm.define "shark1" do |conf|
    conf.vm.network :private_network, ip: "10.10.10.11"
    conf.vm.provider :virtualbox do |vb|
        vb.customize ["modifyvm", :id, "--memory", "256"]
        vb.customize ["modifyvm", :id, "--vram", "1"]
        vb.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]
    end
  end

  config.vm.define "guppy" do |conf|
  	conf.vm.network :private_network, ip: "10.10.10.100"
  end

  config.vm.define "fishtank" do |conf|
    conf.vm.network :private_network, ip: "10.10.10.200"
  end
end