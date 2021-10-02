| key                          | value          | expire |
|------------------------------|----------------|--------|
| userYYYYMMDD                 | counter userId | 24h    |
| uuid (access token)          | userId         | 30m    |
| uuid++userId (refresh token) | userId         | 7d     |