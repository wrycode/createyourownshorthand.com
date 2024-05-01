// get all radio buttons
var radios = document.getElementsByName('feature_choice');
// textarea element
var textBox = document.getElementById('new_features'); 

// function to check radio button change
var radioChange = function() {
    if (this.value == "something else") {
	// enable textarea and clear its value
	textBox.removeAttribute('disabled'); 
	textBox.value = '';
    } else {
	// disable textarea and clear its value
	textBox.setAttribute('disabled', 'disabled'); 
	textBox.value = '';
    }
}

// loop over each radio button
for (var i = 0, length = radios.length; i < length; i++) {
    radios[i].addEventListener('change', radioChange);
}

document.querySelectorAll("li").forEach(item => {
    item.addEventListener("click", event => {
	let radioButton = event.target.querySelector("input[type='radio']");
	if (radioButton) {
	    radioButton.checked = true;
	}
    });
});
function formDataAsObject(form) {
    let formData = new FormData(form);
    let formObj = {};

    for(let [key, value] of formData.entries()) {
	formObj[key] = value;
    }

    return formObj;
}

document.querySelector('form').addEventListener('submit', (event) => {
    // Stop the form from submitting the old way
    event.preventDefault();

    let form = document.getElementById("survey_form");
    let formValues = formDataAsObject(form); // Now you have a form values as an object
    console.log(formValues)

    fetch("https://submit-form.com/VGuU4G2Hb", {
	method: "POST",
	headers: {
	    "Content-Type": "application/json",
	    Accept: "application/json",
	},
	body: JSON.stringify(formValues)
    })
	.then(function (response) {
	    if(response.ok) {
		// Hide the form
		document.querySelector('form').style.display = 'none';
    
		// Display thank you message
		let thankYouMessage = document.createElement('p');
		thankYouMessage.innerText = 'Thank you for your feedback!';
		document.getElementById("main-content").append(thankYouMessage)
		    // document.body.appendChild(thankYouMessage);

	    }
	    console.log(response);
	})
	.catch(function (error) {
	    console.error(error);
	});
})


