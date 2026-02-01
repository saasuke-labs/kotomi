#!/usr/bin/env python3

"""
Example JWT Token Generator for Kotomi Authentication

This script demonstrates how to generate JWT tokens for use with Kotomi's
external authentication mode (Phase 1).

Usage:
    pip install pyjwt
    python generate_jwt.py
"""

import jwt
import os
import time
import json
from datetime import datetime, timedelta

# Configuration - Replace with your actual values
CONFIG = {
    # For HMAC (symmetric key)
    'secret': os.getenv('KOTOMI_JWT_SECRET', 'your-secret-key-min-32-chars-long'),
    
    # For RSA (asymmetric key) - use your private key file
    # 'private_key_path': './path/to/private-key.pem',
    
    # JWT settings
    'issuer': 'https://example.com',
    'audience': 'kotomi',
    'algorithm': 'HS256',  # Use 'RS256' for RSA
}

# User information to encode in the token
user_info = {
    'id': 'user-12345',
    'name': 'John Doe',
    'email': 'john@example.com',
    'avatar_url': 'https://example.com/avatar.jpg',
    'profile_url': 'https://example.com/profile/john',
    'verified': True,
    'roles': ['user', 'contributor'],
}


def generate_hmac_token():
    """Generate a JWT token using HMAC (symmetric key)"""
    now = int(time.time())
    
    payload = {
        'iss': CONFIG['issuer'],
        'sub': user_info['id'],
        'aud': CONFIG['audience'],
        'exp': now + (60 * 60),  # 1 hour expiration
        'iat': now,
        'kotomi_user': user_info,
    }
    
    token = jwt.encode(payload, CONFIG['secret'], algorithm=CONFIG['algorithm'])
    
    return token


def generate_rsa_token(private_key):
    """Generate a JWT token using RSA (asymmetric key)"""
    now = int(time.time())
    
    payload = {
        'iss': CONFIG['issuer'],
        'sub': user_info['id'],
        'aud': CONFIG['audience'],
        'exp': now + (60 * 60),  # 1 hour expiration
        'iat': now,
        'kotomi_user': user_info,
    }
    
    token = jwt.encode(payload, private_key, algorithm='RS256')
    
    return token


def main():
    print('=== Kotomi JWT Token Generator ===\n')
    
    try:
        # Generate HMAC token
        hmac_token = generate_hmac_token()
        print('HMAC Token (HS256):')
        print(hmac_token)
        print()
        
        # Decode to show payload (for debugging)
        decoded = jwt.decode(hmac_token, CONFIG['secret'], 
                           algorithms=[CONFIG['algorithm']],
                           audience=CONFIG['audience'],
                           issuer=CONFIG['issuer'])
        print('Decoded Payload:')
        print(json.dumps(decoded, indent=2))
        print()
        
        # Usage example
        print('=== Usage Example ===')
        print('Include this token in your API requests:')
        print()
        print(f'curl -X POST https://kotomi.example.com/api/v1/site/{{siteId}}/page/{{pageId}}/comments \\')
        print(f'  -H "Authorization: Bearer {hmac_token}" \\')
        print(f'  -H "Content-Type: application/json" \\')
        print(f'  -d \'{{\"text\": \"This is my comment\"}}\'')
        print()
        
        # Python example
        print('=== Python Example ===')
        print('import requests')
        print()
        print('url = "https://kotomi.example.com/api/v1/site/{siteId}/page/{pageId}/comments"')
        print('headers = {')
        print(f'    "Authorization": "Bearer {hmac_token}",')
        print('    "Content-Type": "application/json"')
        print('}')
        print('data = {"text": "This is my comment"}')
        print()
        print('response = requests.post(url, headers=headers, json=data)')
        print('print(response.json())')
        print()
        
        # Configuration instructions
        print('=== Configuration Instructions ===')
        print('1. Log in to Kotomi admin panel')
        print('2. Navigate to Site Settings → Authentication')
        print('3. Configure authentication:')
        print(f'   - Auth Mode: external')
        print(f'   - Validation Type: hmac')
        print(f'   - JWT Secret: {CONFIG["secret"]}')
        print(f'   - JWT Issuer: {CONFIG["issuer"]}')
        print(f'   - JWT Audience: {CONFIG["audience"]}')
        print()
        
    except Exception as e:
        print(f'Error generating token: {e}')
        return 1
    
    return 0


if __name__ == '__main__':
    exit(main())
