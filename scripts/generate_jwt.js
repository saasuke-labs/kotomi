#!/usr/bin/env node

/**
 * Example JWT Token Generator for Kotomi Authentication
 * 
 * This script demonstrates how to generate JWT tokens for use with Kotomi's
 * external authentication mode (Phase 1).
 * 
 * Usage:
 *   npm install jsonwebtoken
 *   node generate_jwt.js
 */

const jwt = require('jsonwebtoken');

// Configuration - Replace with your actual values
const CONFIG = {
  // For HMAC (symmetric key)
  secret: process.env.KOTOMI_JWT_SECRET || 'your-secret-key-min-32-chars-long',
  
  // For RSA (asymmetric key) - use your private key file
  // privateKeyPath: './path/to/private-key.pem',
  
  // JWT settings
  issuer: 'https://example.com',
  audience: 'kotomi',
  algorithm: 'HS256', // Use 'RS256' for RSA
};

// User information to encode in the token
const userInfo = {
  id: 'user-12345',
  name: 'John Doe',
  email: 'john@example.com',
  avatar_url: 'https://example.com/avatar.jpg',
  profile_url: 'https://example.com/profile/john',
  verified: true,
  roles: ['user', 'contributor'],
};

/**
 * Generate a JWT token using HMAC (symmetric key)
 */
function generateHMACToken() {
  const payload = {
    iss: CONFIG.issuer,
    sub: userInfo.id,
    aud: CONFIG.audience,
    exp: Math.floor(Date.now() / 1000) + (60 * 60), // 1 hour expiration
    iat: Math.floor(Date.now() / 1000),
    kotomi_user: userInfo,
  };

  const token = jwt.sign(payload, CONFIG.secret, {
    algorithm: CONFIG.algorithm,
  });

  return token;
}

/**
 * Generate a JWT token using RSA (asymmetric key)
 */
function generateRSAToken(privateKey) {
  const payload = {
    iss: CONFIG.issuer,
    sub: userInfo.id,
    aud: CONFIG.audience,
    exp: Math.floor(Date.now() / 1000) + (60 * 60), // 1 hour expiration
    iat: Math.floor(Date.now() / 1000),
    kotomi_user: userInfo,
  };

  const token = jwt.sign(payload, privateKey, {
    algorithm: 'RS256',
  });

  return token;
}

// Main execution
try {
  console.log('=== Kotomi JWT Token Generator ===\n');
  
  // Generate HMAC token
  const hmacToken = generateHMACToken();
  console.log('HMAC Token (HS256):');
  console.log(hmacToken);
  console.log('');
  
  // Decode to show payload (for debugging)
  const decoded = jwt.decode(hmacToken);
  console.log('Decoded Payload:');
  console.log(JSON.stringify(decoded, null, 2));
  console.log('');
  
  // Usage example
  console.log('=== Usage Example ===');
  console.log('Include this token in your API requests:');
  console.log('');
  console.log('curl -X POST https://kotomi.example.com/api/v1/site/{siteId}/page/{pageId}/comments \\');
  console.log('  -H "Authorization: Bearer ' + hmacToken + '" \\');
  console.log('  -H "Content-Type: application/json" \\');
  console.log('  -d \'{"text": "This is my comment"}\'');
  console.log('');
  
  // Configuration instructions
  console.log('=== Configuration Instructions ===');
  console.log('1. Log in to Kotomi admin panel');
  console.log('2. Navigate to Site Settings → Authentication');
  console.log('3. Configure authentication:');
  console.log('   - Auth Mode: external');
  console.log('   - Validation Type: hmac');
  console.log('   - JWT Secret: ' + CONFIG.secret);
  console.log('   - JWT Issuer: ' + CONFIG.issuer);
  console.log('   - JWT Audience: ' + CONFIG.audience);
  console.log('');
  
} catch (error) {
  console.error('Error generating token:', error.message);
  process.exit(1);
}
