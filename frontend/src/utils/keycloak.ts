import Keycloak from 'keycloak-js'
import { tokenManager } from './tokenManager'

const keycloakUrl = import.meta.env.VITE_KEYCLOAK_URL || 'http://localhost:8090'
const keycloakRealm = import.meta.env.VITE_KEYCLOAK_REALM || 'multi-agent'
const keycloakClientId = import.meta.env.VITE_KEYCLOAK_CLIENT_ID || 'ecommerce-system-12345'

export const keycloak = new Keycloak({
  url: keycloakUrl,
  realm: keycloakRealm,
  clientId: keycloakClientId,
})

export const initKeycloak = (onSuccess: () => void) => {
  keycloak
    .init({
      onLoad: 'check-sso',
      pkceMethod: 'S256',
      checkLoginIframe: false,
    })
    .then((authenticated) => {
      if (authenticated && keycloak.token) {
        tokenManager.setToken(keycloak.token)
      } else {
        tokenManager.removeToken()
      }
      onSuccess()
    })
    .catch((err) => {
      console.error('Failed to initialize Keycloak:', err)
      onSuccess() // Fallback to let React render anyway
    })

  keycloak.onTokenExpired = () => {
    keycloak
      .updateToken(30) // check xem access token còn hạn dưới 30s không
      .then((refreshed) => {
        if (refreshed && keycloak.token) {
          tokenManager.setToken(keycloak.token)
        }
      })
      .catch((err) => {
        console.error('Failed to refresh Keycloak token:', err)
        tokenManager.removeToken()
      })
  }
}
