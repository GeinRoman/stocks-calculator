# Stocks calculator project

## Description

This project is designed to help me manage my stocks. Fractions of each company's stocks in relation to total portfolio cost change over some period, so next time I want to deposit money (or just rebalance portfolio) I need to know which stocks to buy or sell to preserve original ratios.
So app has functionality of adding groups of stocks that represents industrial fields. User is able to set ratios between these groups and ratios between individual stocks inside each group. Using whese ratios and current stock prices program will present information on what to buy/sell to rebalance portfolio.

Also I want to be able to:
* see information on my portfolio (only stocks for now, support for bonds etc. may be later)
* add or remove record from portfolio
* see price changes and returns of portfolio or stocks

## Stucture of the project

Project should be devided to 2 parts:
1. cli application that user will interact with
2. backend service that stores stocks info in db or gets it from the internet and deliveres it to users

This division makes it possible to create an account system, so I can access same database from multiple devices (and even have multiple accounts)

## CLI application

To show list of stocks in portfolio
        stcalc list shows
        flags: 
        * -s --short    consice version of a command
        * -v --verbose  all information about stocks
