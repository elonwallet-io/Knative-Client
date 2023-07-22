![ElonWallet](https://github.com/elonwallet-io/dev-deploy/assets/57064670/54b91e8d-ebf8-43d3-a453-880d292c1f9e)
# Knative-Client, also known as the Deployer.
The backend of Elonwallet is able to call two API Endpoints of the Knative-Client to either deploy Wallet-Services or remove them.
Every Wallet-Service gets additionally a Volume mounted for persistent storage by the Knative-Client. See [here for documentation](https://docs.elonwallet.io/development/architecture/deployer).

# Future Work:
- Deploy the Vault and configure it.
- Create a private key for the wildcard-certificate. Issue a CSR and provide an endpoint for the CSR to be received by an admin  
