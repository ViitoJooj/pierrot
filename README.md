```bash
                                              ======                     
                                            ==========                   
                    -------     =======     ============      --------    
                  ---------   ===========  =============+     ----------  
                  ------     ====================    ==++         ------  
                ------     ====== ==============     +=-:         -----  
                ------    ++++    ==============      :::         -----  
                ------   ::--     ===============                 -----  
                ------    ::::     ===============                  ------
            ---------            ================                  ---------
            ---------            ======+++========                 ---------
                ------           ======+*+*+==++===                 ------
                ------          ++*+++*=:-++++*+++                -----  
                ------         -+++=-:::::::-=+++=:               -----  
                ------        ::-=-::-=-:-=-::-==-::              -----  
                ------        ::=*+-:=+=:-++-:-++-::              -----  
                  ------       ::::::::::::::::::::::             ------  
                  ---------                                   --------- 
```

<h1 align="center">Pierrot <img src="./imgs/pierrot.svg" width="20px" height="20px"></h1>

### [Read the docs →](./docs/documentation.md)
## What is pierrot ?
Pierrot is a JavaScript-free framework built in Golang. It functions as a compiler that understands the `.pierrot` language, but also allows the use of TypeScript when needed. Its goal is to offer a simple and flexible way to develop web applications, without imposing limitations on the developer.

## how install ?
```bash
go install github.com/ViitoJooj/pierrot/cmd/pierrot@latest
```
This command will install the Pierrot binary in ~/.pierrot/bin/pierrot.exe. After installation, you will be able to use the Pierrot command-line commands, shown in the example below.

Alternatively, you can download the latest release by clicking <a href="https://github.com/ViitoJooj/pierrot/releases/">here</a>.

We recommend creating a `.pierrot` directory in your user folder and placing the binary inside `.pierrot/bin`. After that, add the directory to your environment variables so the `pierrot` command can be accessed from anywhere in your terminal.

## Simple usage:
To start a project with the Pierrot architecture, you need to run the following command:
```bash
pierrot init <project-name>
```
This command will create your website/app folders in the following format:

```txt
project-name/
├── src/
│   ├── assets/                     # static files
│   │   ├── robots.txt              # search engine instructions
│   │   └── favicon.ico             # website icon
│   ├── components/                 # reusable components
│   │   └── header/
│   │       ├── script.ts           # component logic
│   │       ├── styles.css          # component styles
│   │       └── index.pierrot       # component template
│   ├── pages/                      # application routes
│   │   ├── errors/                 # "*" router fallback
│   │   │   ├── script.ts           # error page logic
│   │   │   ├── styles.css          # error page styles
│   │   │   └── index.pierrot       # fallback page template status (404)
│   │   └── home/                   # the router for "/home" and "/"
│   │       ├── script.ts           # home page logic
│   │       ├── styles.css          # home page styles
│   │       └── index.pierrot       # page template rendered by Pierrot
│   ├── globals.css                 # global styles and variables
│   └── main.pierrot                # application entry point
└── setting.pierrot.json            # Pierrot configuration file
```
### how start ?
use this command for development mode:
```bash
cd <project-name>
pierrot dev
```
and to build use:
```bash
cd <project-name>
pierrot build
```
Remember that these two commands need to be executed using settings.pierrot.json.


## Contributing

Refer to the [Project > Contributing](./CONTRIBUTING.md) guide to start contributing to Pierrot.

## License

Refer to the [Project > License](./LICENSE.md) page for information about Pierrot's licensing.