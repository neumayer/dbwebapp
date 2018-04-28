control 'container' do
  impact 0.5
  describe docker_container('dbwebapp') do
    it { should exist }
    it { should be_running }
    its('repo') { should eq 'neumayer/dbwebapp' }
    its('ports') { should eq '0.0.0.0:8080->8080/tcp' }
    its('command') { should match '/dbwebapp' }
  end
end       
