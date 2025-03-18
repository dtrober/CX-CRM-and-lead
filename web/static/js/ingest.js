import "https://unpkg.com/maplibre-gl@4.7.1/dist/maplibre-gl.js";

const middleOfUSA = [-100, 40];


async function init() {
    const map = new maplibregl.Map({
      style: "/static/mapStyle/dark.json",
      // style: "https://tiles.openfreemap.org/styles/liberty",
      center: middleOfUSA,
      zoom: 2,
      container: "map",
    });  
}


document.addEventListener('DOMContentLoaded', function() {
    const form = document.getElementById('business-search-form');
    const validateBtn = document.getElementById('validate-btn');
    const searchBtn = document.getElementById('search-btn');
    const validationMessage = document.getElementById('validation-message');
    const searchResults = document.getElementById('search-results');

    // Handle validation button click
    validateBtn.addEventListener('click', function() {
        validateForm();
    });

    // Handle form submission
    form.addEventListener('submit', function(e) {
        e.preventDefault();
        
        // First validate
        if (validateForm()) {
            // Then display fake results (no API call)
            displayFakeResults();
        }
    });

    // Validate form fields
    function validateForm() {
        const location = document.getElementById('location').value.trim();
        const businessType = document.getElementById('businessType').value.trim();
        
        // Reset validation message
        validationMessage.className = 'message-container';
        validationMessage.textContent = '';
        
        // Basic validation
        if (!location) {
            showError('Please enter a zip code or city');
            return false;
        }
        
        if (!businessType) {
            showError('Please enter a business type');
            return false;
        }
        
        // Validate zip code format if it appears to be a zip code
        const zipRegex = /^\d{5}(-\d{4})?$/;
        if (/^\d+$/.test(location) && !zipRegex.test(location)) {
            showError('Please enter a valid 5-digit zip code');
            return false;
        }
        
        // If we made it here, validation passed
        showSuccess('Validation successful! You can now search.');
        return true;
    }

    // Function to display fake business results
    function displayFakeResults() {
        const location = document.getElementById('location').value.trim();
        const businessType = document.getElementById('businessType').value.trim();
        
        // Clear previous results
        searchResults.innerHTML = '';
        
        // Create a header
        const resultsHeader = document.createElement('h2');
        resultsHeader.textContent = 'Search Results';
        searchResults.appendChild(resultsHeader);
        
        // Generate 3 fake businesses
        const businesses = [
            {
                name: `${businessType} Express`,
                address: "123 Main Street",
                city: isZipCode(location) ? "Anytown" : location,
                state: "CA",
                zip: isZipCode(location) ? location : "90210",
                type: businessType,
                phone: "555-123-4567"
            },
            {
                name: `${businessType} Professionals`,
                address: "456 Oak Avenue",
                city: isZipCode(location) ? "Somewhere" : location,
                state: "CA",
                zip: isZipCode(location) ? location : "90211",
                type: businessType,
                phone: "555-987-6543"
            },
            {
                name: `Premium ${businessType}`,
                address: "789 Elm Boulevard",
                city: isZipCode(location) ? "Nowhere" : location,
                state: "CA",
                zip: isZipCode(location) ? location : "90212",
                type: businessType,
                phone: "555-789-0123"
            }
        ];
        
        // Create a container for the business cards
        const businessList = document.createElement('div');
        businessList.className = 'business-list';
        
        // Add each business
        businesses.forEach(business => {
            const businessCard = document.createElement('div');
            businessCard.className = 'business-card';
            businessCard.innerHTML = `
                <h3>${business.name}</h3>
                <p>${business.address}</p>
                <p>${business.city}, ${business.state} ${business.zip}</p>
                <p>Type: ${business.type}</p>
                <p>Phone: ${business.phone}</p>
            `;
            businessList.appendChild(businessCard);
        });
        
        searchResults.appendChild(businessList);
    }
    
    // Helper function to check if a string is a zip code
    function isZipCode(str) {
        return /^\d{5}(-\d{4})?$/.test(str);
    }

    // Helper function to show error message
    function showError(message) {
        validationMessage.className = 'message-container error';
        validationMessage.textContent = message;
    }

    // Helper function to show success message
    function showSuccess(message) {
        validationMessage.className = 'message-container success';
        validationMessage.textContent = message;
    }
});

init(); 