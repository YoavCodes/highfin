Vagrant::Config.run do |config|
 
  config.vm.box = "bytepark/debian-7.4"
  config.vm.network :hostonly, "10.10.10.50"
  config.vm.provision :shell, :path => "vagrant_setup.sh"

  config.vm.customize ["modifyvm", :id, "--memory", "160"]
  config.vm.customize ["modifyvm", :id, "--vram", "1"]
  config.vm.customize ["modifyvm", :id, "--cpuexecutioncap", "60"]

end
