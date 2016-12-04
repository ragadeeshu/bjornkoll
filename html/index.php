<?php
$response;
$body_text = "";
exec("python3 ../bjornkoll.py", $response);
$size = 40;
$scoreline = array_shift($response);
$split = explode ( "<br>", $scoreline);
$last_word;
if( sizeof($split) >1){
  $body_text = $body_text . "<p style=\"font-size:${size}px\">$split[0]</p>";
  $body_text = $body_text . "<p id='score' style=\"font-size:${size}px\">$split[1]</p>";
  $last_word = explode ( " ", $split[1])[2];
} else {
  $body_text = $body_text . "<p id='score' style=\"font-size:${size}px\">$split[0]</p>";
  $last_word = explode ( " ", $split[0])[2];
}
$size += 30;
for ($i = 0; $i < count($response); ++$i) {
  $body_text = $body_text . "<pre style=\"font-size:${size}px;\">";
  $body_text = $body_text . "<span>$response[$i]</span>";
  $i++;
  $body_text = $body_text . "<span class=\"shh\">$response[$i]</span>";
  $body_text = $body_text . "</pre>\n";
  $size *= 0.95;
}

?>
<!DOCTYPE html>
<html>
<style> a:link {color:#FFE14B; background-color:transparent;
text-decoration:none} a:visited {color:#FFE14B; background-color:transparent;
text-decoration:none} a:hover {color:#b09100; background-color:transparent;
text-decoration:none} a:active {color:#b09100; background-color:transparent;
text-decoration:none}
body
{
background:#222;
color:#D0B11B;
text-align:center;
}
pre,p
{

margin:0px;
font-family: 'Courier New', Courier, monospace;
}
.shh, pre:hover span { display: none }

pre:hover .shh { display: inline }

<?php if(intval($last_word)>=100){?>
#score {

    -webkit-animation: colorRotate 6s linear 0s infinite;
     -moz-animation: colorRotate 6s linear 0s infinite;
       -o-animation: colorRotate 6s linear 0s infinite;
          animation: colorRotate 6s linear 0s infinite;

}

@keyframes colorRotate {
    from {
        color: rgb(255, 0, 0);
    }
    16.6% {
        color: rgb(255, 0, 255);
    }
    33.3% {
        color: rgb(0, 0, 255);
    }
    50% {
        color: rgb(0, 255, 255);
    }
    66.6% {
        color: rgb(0, 255, 0);
    }
    83.3% {
        color: rgb(255, 255, 0);
    }
    to {
        color: rgb(255, 0, 0);
    }
<?php
}
?>

</style>
<head>
  <title>Deeshu's bear watcher</title>
</head>
<body>
  <h1>Update statistics for <a href="http://www.thebearden.se/">thebearden.se</a></h1>
  <?=$body_text?>
 </body>
 </html>
