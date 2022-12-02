output = {
    global={
        application={
            links={
                {	description= "console web ui",
                     url=Var.Values.console_dns
                }
            }
        }
    },

    enabled=true,
    ingressClass="nginx",
    replicaCount=2,
    provider=Var.Provider,
    license=Var.License,
    ingress={
            annotations={
                "kubernetes.io/tls-acme: 'true'",
                "cert-manager.io/cluster-issuer: letsencrypt-prod",
                "nginx.ingress.kubernetes.io/affinity: cookie",
                "nginx.ingress.kubernetes.io/force-ssl-redirect: 'true'",
                "nginx.ingress.kubernetes.io/proxy-read-timeout: '3600'",
                "nginx.ingress.kubernetes.io/proxy-send-timeout: '3600'",
                "nginx.ingress.kubernetes.io/session-cookie-path: /socket",
            },
            console_dns=Var.Values.console_dns
    },
    serviceAccount= {
            create=true,
            annotations="eks.amazonaws.com/role-arn: arn:aws:iam::" .. Var.Project .. ":role/" ..Var.Cluster .. "-console"

    },
    secrets={
            jwt=dedupe(Var, "console.secrets.jwt", randAlphaNum(20)),
            admin_name=Var.Values.admin_name,
            admin_email=dedupe(Var, "console.secrets.admin_email", default("someone@example.com", Var.Config.Email)),
            admin_password=dedupe(Var, "console.secrets.admin_password", randAlphaNum(20)),
            cluster_name=Var.Cluster,
            erlang=dedupe(Var, "console.secrets.erlang", randAlphaNum(14)),
            id_rsa=ternary(Var.Values.private_key, dedupe(Var, "console.secrets.id_rsa", ""), hasKey (Var.Values, "private_key")),
            id_rsa_pub=ternary(Var.Values.public_key, dedupe(Var, "console.secrets.id_rsa_pub", ""), hasKey(Var.Values, "public_key")),
            ssh_passphrase=ternary(Var.Values.passphrase, dedupe(Var, "console.secrets.ssh_passphrase", ""), hasKey(Var.Values, "passphrase")),
            git_access_token=ternary(Var.Values.access_token, dedupe(Var, "console.secrets.git_access_token", ""), hasKey(Var.Values, "access_token")),
            git_user=default("console", Var.Values.git_user),
            git_email=default("console@plural.sh", Var.Values.git_email),
            git_url="",
            repo_root="",
            branch_name="",
            config="",
            key="",
    }
}

if Var.Provider == "kind" then
    output.ingress.annotations = {
        "external-dns.alpha.kubernetes.io/target: '127.0.0.1'"
    }
    output.replicaCount=1
end

if Var.Provider == "google" then
    output.serviceAccount.create = false
end

if Var.Provider == "azure" then
    output.podLabels={
        "aadpodidbinding: console"
    }
    output.consoleIdentityId=importValue("Terraform", "console_msi_id")
    output.consoleIdentityClientId=importValue("Terraform", "console_msi_client_id")

    output.extraEnv={
        {
            name="ARM_USE_MSI",
            value = true

        },
        {
            name="ARM_SUBSCRIPTION_ID",
            value=Var.Context.SubscriptionId
        },
        {
            name="ARM_TENANT_ID",
            value= Var.Context.TenantId
        }
    }

end

if Var.OIDC ~= nil then
    output.secrets.plural_client_id=Var.OIDC.ClientId
    output.secrets.plural_client_secret=Var.OIDC.ClientSecret
end

if Var.Values.is_demo then
    output.secrets.is_demo=Var.Values.is_demo
end

if Var.Values.console_dns then
    local gitUrl=dig("console", "secrets", "git_url", "default", Var)
    local identity=pathJoin(repoRoot(), ".plural-crypt", "identity")
    if gitUrl == "default" or gitUrl == "" then
        output.secrets.git_url=repoUrl()
    else
        output.secrets.git_url=gitUrl
    end

    --output.secrets.repo_root=repoName()
    output.secrets.branch_name=branchName()
    output.secrets.config=readFile(pathJoin(homeDir(),".plural","config.yml"))

    if fileExists(identity) then
        output.secrets.identity=readFile(identity)
    elseif dig("console", "secrets", "identity", "default", Var) ~= "default" then
        output.secrets.identity= Var.console.secrets.identity
    else
        output.secrets.key=readFile(pathJoin(homeDir(), ".plural", "key"))
    end
end
