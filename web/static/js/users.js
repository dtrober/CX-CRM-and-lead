document.addEventListener('DOMContentLoaded', function() {

    fetchUsers();
    

    const form = document.getElementById('create-user-form');
    if (form) {
        form.addEventListener('submit', function(e) {
            e.preventDefault();
            createUser();
        });
    }
});


function fetchUsers() {
    const userList = document.getElementById('user-list');
    
    fetch('/api/v1/users')
        .then(response => {
            if (!response.ok) {
                throw new Error('Failed to fetch users');
            }
            return response.json();
        })
        .then(data => {
            userList.innerHTML = '';
            
            if (data.length === 0) {
                userList.innerHTML = '<p>No users found.</p>';
                return;
            }
            
            data.forEach(user => {
                const userElement = document.createElement('div');
                userElement.className = 'user-card';
                userElement.innerHTML = `
                    <h3>${user.name}</h3>
                    <p>Email: ${user.email}</p>
                    <p>Created: ${new Date(user.created_at).toLocaleDateString()}</p>
                `;
                userList.appendChild(userElement);
            });
        })
        .catch(error => {
            console.error('Error:', error);
            userList.innerHTML = `<p>Error loading users: ${error.message}</p>`;
        });
}


function createUser() {
    const name = document.getElementById('name').value;
    const email = document.getElementById('email').value;
    
    fetch('/api/v1/users', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
        },
        body: JSON.stringify({
            name: name,
            email: email
        })
    })
    .then(response => {
        if (!response.ok) {
            throw new Error('Failed to create user');
        }
        return response.json();
    })
    .then(data => {
        // Clear form
        document.getElementById('name').value = '';
        document.getElementById('email').value = '';
        
        // Reload user list
        fetchUsers();
        
        alert('User created successfully!');
    })
    .catch(error => {
        console.error('Error:', error);
        alert(`Error creating user: ${error.message}`);
    });
}