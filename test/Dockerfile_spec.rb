require "docker"
require "serverspec"

describe "Dockerfile" do
  before(:all) do
    @image = Docker::Image.build_from_dir(".")
    #@image = Docker::Image.get("neumayer/dbwebapp")

    set :os, family: :redhat
    set :backend, :docker
    set :docker_image, @image.id
    set :docker_container_create_options, { "Entrypoint" => ["sh"] }
  end

  it "exposes correct ports" do
    expect(@image.json["ContainerConfig"]["ExposedPorts"]).to include("8080/tcp")
  end

  it "sets /dbwebapp as entrypoint" do
    expect(@image.json["Config"]["Entrypoint"][0]).to eq("/dbwebapp")
  end

  it "contains binary" do
    expect(file("/dbwebapp")).to be_file
  end
end
