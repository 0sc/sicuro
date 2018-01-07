# SicuroCI

SicuroCI is an exploration of the various pieces that interplay to form a CI server. The results form the basis of [this article](https://medium.com/p/sicuro-ci-2f40ba138233) which discusses each of the various pieces; how, why, when and where they come in, in the entire CI server life cycle.

## Dependencies
To run the app, you'll need to have installed
* [Docker & Docker composer](https://docs.docker.com/engine/installation/) - tests are run in docker containers
* [Golang](https://golang.org/doc/install) - the app is written in Go
* [Glide](http://glide.sh/) - package manager for Go

You would also need to setup, Github OAuth App, Webhook and SSH Keys.

## Running the application
If you haven't checked out [the article linked above](https://medium.com/p/sicuro-ci-2f40ba138233), you would want to pause here and go give it a [read](https://medium.com/p/sicuro-ci-2f40ba138233). It contains explanation on how to obtain necessary credentials you would require. These include
* GITHUB_CLIENT_ID
* GITHUB_CLIENT_SECRET
* GITHUB_WEBHOOK_SECRET

Follow these steps to run the app
* Create a folder `.ssh` in the [`ci` folder](./ci). Inside the `.ssh` subfolder add the github SSH keys you created above; these should be two files `id_rsa` and `id_rsa.pub` which are the private and public keys respectively.
* Update the env section of the [Makefile](./Makefile) with the relevant information
* Ensure you have docker and docker-compose running
* Within the root of the app, execute: `make start`

The app should now be available at `localhost:PORT`. _(the port would be what you set it to in the [Makefile](./Makefile). Default value is `8080`)_

To test the github webhook locally, you'll need to setup a URL that tunnels requests over the internet to the server running on your machine. A good tool of choice is Ngrok, easy to setup and use. See instructions [here](https://ngrok.com/download)

<img width="556" alt="screen shot 2018-01-07 at 5 32 46 pm" src="https://user-images.githubusercontent.com/11221027/34655897-b7c220d0-f411-11e7-9eea-fe965e323da7.png">

Once you have the tunnel setup (and running), you should be assigned a URL that would tunnel requests to your local server, for example, https://example.ngrok.io. Head over to your OAuth dashboard on Github and set the callback URL to https://example.ngrok.io/gh/callback _(replace https://example.ngrok.io)_ with your ngrok URL.

You can now browse the app with the ngrok URL and try out the webhook callback.
<img width="421" alt="screen shot 2018-01-07 at 11 39 38 pm" src="https://user-images.githubusercontent.com/11221027/34655358-5b37e8a8-f408-11e7-9742-e254f6d1e51b.png">
<img width="570" alt="screen shot 2018-01-07 at 11 50 03 pm" src="https://user-images.githubusercontent.com/11221027/34655359-5b5f76a2-f408-11e7-81e6-46d63b0c940b.png">
<img width="824" alt="screen shot 2018-01-07 at 11 53 21 pm" src="https://user-images.githubusercontent.com/11221027/34655360-5b886666-f408-11e7-8340-7c2f942a42fa.png">

## Contributing

Bug reports and pull requests are welcome on GitHub at https://github.com/0sc/sicuro. This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](http://contributor-covenant.org) code of conduct.

## License

The app is available as open source under the terms of the [MIT License](https://opensource.org/licenses/MIT).