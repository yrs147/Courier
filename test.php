<!DOCTYPE html>
<html>
<head>
	<title>Execute Go program</title>
</head>
<body>
	<button id="execute-btn">Execute Go program</button>
	<div id="result"></div>

	<script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
	<script>
		$(document).ready(function() {
			$("#execute-btn").click(function() {
				$.ajax({
					url: "execute.php",
					type: "GET",
					success: function(response) {
						$("#result").html(response);
					},
					error: function(jqXHR, textStatus, errorThrown) {
						alert("Error: " + textStatus + " - " + errorThrown);
					}
				});
			});
		});
	</script>
</body>
</html>
