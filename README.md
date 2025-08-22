# Gator

## Installation
In order to run Gator, you need to have Go installed on your machine. You can download Go from the official website: https://golang.org/dl/

After installing Go, you can clone the Gator repository using the following command:

```
git clone https://github.com/jzetterman/gator.git
```

Then, navigate to the cloned directory and run the following command to install Gator:

```
go install
```

Next you need to install Postgres and create a database for Gator to use. You can download Postgres from the official website: https://www.postgresql.org/download/

After installing Postgres, you can create a database using the following command:

```
createdb gator
```

Finally, you need to configure Gator to use the database. You can do this by creating a file named `.gatorconfig.json` in the root of your home directory and adding the following lines:

```
{"db_url":"postgres://postgres:@localhost:5432/gator?sslmode=disable"}
```

## Commands

To create a new user, run the following command:

```
gator register <username>
```

To log into a user run the following command:

```
gator login <username>
```

To get a list of users, run the following command:

```
gator users
```

To add an RSS feed, run the following command:

```
gator addfeed <feed_name> <url>
```

To follow a feed, run the following command:

```
gator follow <feed_name>
```
