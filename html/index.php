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

</style>
<head>
  <title>Deeshu's bear watcher</title> </head> <body>
  <h1>Update statistics for <a href="http://www.thebearden.se/">thebearden.se</a></h1>
  <?php
  $response;
  exec("python3 ../bjornkoll.py", $response);
  $size = 40;
  $scoreline = array_shift($response);
  print "<p style=\"font-size:${size}px\">$scoreline</p>";
  $size += 30;
  for ($i = 0; $i < count($response); ++$i) {
    print "<pre style=\"font-size:${size}px;\">";
    print "<span>$response[$i]</span>";
    $i++;
    print "<span class=\"shh\">$response[$i]</span>";
    print "</pre>";
    $size *= 0.95;
  }
  ?> </body> </html>
