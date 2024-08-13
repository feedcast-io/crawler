# Feedcast Crawler

## Send request to Google Function 

```
POST https://europe-west9-feedcast-2023.cloudfunctions.net/web-crawler
{
    domain: "www.orixa-media.com",
    max_pages: uint16, // 0-65536, default = 100
    max_duration: uint8, // 0-256, default = 30
    max_depth: uint8 // 0-256, default = 4
}
```

