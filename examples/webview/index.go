package main

const indexPage = `
<html>
<head>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.4.0/css/bootstrap.css" />
	
</head>
<body>
	<nav class="navbar navbar-expand navbar-light bg-light">
		<a class="navbar-brand" href="#">VGE Webview sample</a>
		<div class="collapse navbar-collapse" id="navbarsExample02">
        <ul class="navbar-nav mr-auto">
          <li class="nav-item active">
            <a class="nav-link" href="#">Home <span class="sr-only">(current)</span></a>
          </li>
        </ul>
      </div>
	</nav>
	<div class="container">
		<div class="mb-2 mt-2">
			<img style="height: 80vh" id="img" src="/getImg?angle=0" />
		</div>
		<div class="row">
			<div class="col-3">
				<input type="range" min="0" max="360" value="0" id="angle" />
				<label> Angle: </label>
				<label id="lAngle">0</label>
			</div>
		</div>
	</div>
	<script type="application/javascript">
	var iAngle = document.getElementById('angle');
	var ilAngle = document.getElementById('lAngle');
	var iImg = document.getElementById('img');
	function angleChanged() {
	    ilAngle.innerText = iAngle.value;
	    iImg.src = '/getImg?angle=' + iAngle.value + '&_t=' + Date.now().toString();
	}	
	iAngle.addEventListener('input', angleChanged);
	</script>
</body>
`
