// Rot13 decoder script
const encodedString = 'Pbatenghyngvbaf ba ohvyqvat n pbqr-rqvgvat ntrag!';

function rot13decode(str) {
  return str.replace(/[a-zA-Z]/g, function(char) {
    // Get the character code
    const charCode = char.charCodeAt(0);
    
    // Handle uppercase letters
    if (charCode >= 65 && charCode <= 90) {
      return String.fromCharCode(((charCode - 65 + 13) % 26) + 65);
    }
    
    // Handle lowercase letters
    if (charCode >= 97 && charCode <= 122) {
      return String.fromCharCode(((charCode - 97 + 13) % 26) + 97);
    }
    
    // Return non-alphabetic characters unchanged
    return char;
  });
}

// Decode and print the message
const decodedMessage = rot13decode(encodedString);
console.log(decodedMessage);