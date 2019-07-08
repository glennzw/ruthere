# Are you there?
    
Are you there (ruthere) is a tiny endpoint to query the response code of a supplied URL. A square image is returned, the dimensions of which are the response code.

For example:  
https://ruthere.co/u/www.google.com/robots.txt will rerturn a 200x200px image.
https://ruthere.co/u/www.google.com/nosuchpage will return a 404x404px image.
https://ruthere.co/u/http://getstatuscode.com/503 will return a 503x503px image.

Non HTTP response codes are also returned for when things go wrong:  
https://ruthere.co/u/sdasdasdsa will return a 50x50px image (failed to validate URL)
https://ruthere.co/u/nosuchsite24324234.com will return 40x40px image (can't connect)

 A returned image of 30x30px indicates a Panic from this code.

 Redirects are followed.

## Why, tho?
 What's the point? Well, this will allow you to check response codes of remote web sites via JavaScript, e.g:

```
var url = "https://twitter.com";
var newImg = new Image;
newImg.src = "https://ruthere.co/u/" + url;
console.log("Checking " + url)

newImg.onload = function(){
    console.log(this.width);
    if (newImg.height == 200) {
        console.log("Twitter is up!");
    } else {
        console.log("Twitter broke :(");
    }
}
```
 If you don't know why this is useful, you probably don't need it.