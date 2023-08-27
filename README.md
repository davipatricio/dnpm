# dnpm
A simple, fast and efficient Node Package Manager written in Go 🐕⚡

## 🥰 Sponsors
Nobody :(

## 💻 Building from source
Clone the repository:
```sh-session
$ git clone https://github.com/davipatricio/dnpm.git
```

Build the binary for your platform:
```sh-session
$ cd dnpm
$ go build cmd/dnpm.go
```

> If you wish a smaller binary size, compile the program using `go build -ldflags="-w -s" -gcflags=all="-l -B" cmd/dnpm.go`

## 💻 Install from npm
It is recommended to install dnpm through the [Node Package Manager](http://npmjs.org/), which comes with [Node.js](https://nodejs.org) when you install it on your system.

Once you have npm installed you can run the following both to **install** and **upgrade** dnpm:
```sh-session
npm install -g @davipatricio/dnpm
```

## 🕵️ Check installation
Check that dnpm is installed by running:
```sh-session
$ dnpm version
```

## 📚 Usage
Please refer to the [documentation](docs) for more information.

## 📝 License
This project is licensed under the [MIT](LICENSE) license.
