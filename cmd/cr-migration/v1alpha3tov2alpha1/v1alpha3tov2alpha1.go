package v1alpha3tov2alpha1

import (
	"fmt"
	"log/slog"

	migrationutils "github.com/astarte-platform/astartectl/cmd/cr-migration"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// convertCassandraConnectionSpec converts the spec.cassandra.connection section from v1alpha3 to v2alpha1
func convertCassandraConnectionSpec(oldSpec *unstructured.Unstructured) (newConnection *unstructured.Unstructured) {
	newConnection = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldConnection, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cassandra", "connection")
	if err != nil {
		slog.Error("error retrieving cassandra connection spec", "err", err)
	}

	if !found || oldConnection == nil {
		slog.Error("spec.cassandra.connection section is missing or empty in the input CR. Resulting CR will have no cassandra connection spec resulting in a invalid Astarte CR.")
	}

	// The following fields are deep copied from the old connection to the new connection. No changes here
	dc := []string{"poolSize", "sslConfiguration"}
	for _, f := range dc {
		sourcePath := []string{"spec", "cassandra", "connection", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newConnection, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	// spec.cassandra.connection.secret -> spec.cassandra.connection.credentialsSecret
	slog.Warn("spec.cassandra.connection.username is no longer supported in v2alpha1 and will be ignored. You will need to set it in the credentialsSecret.")
	slog.Warn("spec.cassandra.connection.password is no longer supported in v2alpha1 and will be ignored. You will need to set it in the credentialsSecret.")
	slog.Warn("spec.cassandra.connection.autodiscovery is no longer supported in v2alpha1 and will be ignored.")

	sourcePath := []string{"spec", "cassandra", "connection", "secret"}
	destPath := []string{"credentialsSecret"}
	err = migrationutils.CopyIfExists(oldSpec, newConnection, sourcePath, destPath)
	if err != nil {
		slog.Error("error copying secret to credentialsSecret", "err", err)
	}

	// spec.cassandra.connection.nodes conversion
	if nodes := migrationutils.ParseCassandraStrNodes(oldSpec); len(nodes) > 0 {
		if err := unstructured.SetNestedSlice(newConnection.Object, nodes, "nodes"); err != nil {
			slog.Error("error setting cassandra connection nodes", "err", err)
		}
	}

	return newConnection
}

// convertCassandraSpec converts the spec.cassandra section from v1alpha3 to v2alpha1
func convertCassandraSpec(oldSpec *unstructured.Unstructured) (newCassandra *unstructured.Unstructured) {
	newCassandra = &unstructured.Unstructured{Object: map[string]interface{}{}}

	oldCassandra, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cassandra")
	if err != nil {
		slog.Error("error retrieving cassandra spec", "err", err)
		return newCassandra
	}

	if !found || oldCassandra == nil {
		slog.Error("spec.cassandra section is missing or empty in the input CR. Resulting CR will have no cassandra spec resulting in a invalid Astarte CR.")
		return newCassandra
	}

	slog.Info("The following fields are no longer supported and will be ignored if present in the source CR: cassandra.deploy, cassandra.replicas, cassandra.image, cassandra.version, cassandra.storage, cassandra.maxHeapSize, cassandra.heapNewSize, cassandra.resources.")

	// spec.astarteSystemKeyspace is now spec.cassandra.astarteSystemKeyspace
	// Build the inner cassandra object content directly (do not nest a top-level "cassandra" key here)
	err = migrationutils.CopyIfExists(oldSpec, newCassandra, []string{"spec", "astarteSystemKeyspace"}, []string{"astarteSystemKeyspace"})
	if err != nil {
		slog.Error("error copying astarteSystemKeyspace", "err", err)
	}

	// If spec.cassandra.deploy is true, ask the user for cassandra.connection.credentialsSecret details
	deploy, foundDepoly, errDeploy := unstructured.NestedBool(oldSpec.Object, "spec", "cassandra", "deploy")
	if deploy && foundDepoly && errDeploy == nil {
		slog.Error("spec.cassandra.deploy is set to true. With the new CR, you will need to deploy Scylla yourself and configure cassandra.connection accordingly. Not doing so will result in a broken Astarte deployment.")

		cassandraConnectionSecretName := ""
		cassandraConnectionSecretUsernameKey := ""
		cassandraConnectionSecretPasswordKey := ""

		// Since the deploy was ture, we ask user for pointers to the new connection
		fmt.Print("new cassandra.connection.credentialsSecret.name: ")
		_, err = fmt.Scanln(&cassandraConnectionSecretName)
		if err != nil {
			slog.Error("error reading cassandra connection credentialsSecret name from input", "err", err)
		}

		fmt.Print("new cassandra.connection.credentialsSecret.usernameKey: ")
		_, err = fmt.Scanln(&cassandraConnectionSecretUsernameKey)
		if err != nil {
			slog.Error("error reading cassandra connection credentialsSecret usernameKey from input", "err", err)
		}

		fmt.Print("new cassandra.connection.credentialsSecret.passwordKey: ")
		_, err = fmt.Scanln(&cassandraConnectionSecretPasswordKey)
		if err != nil {
			slog.Error("error reading cassandra connection credentialsSecret passwordKey from input", "err", err)
		}

		slog.Info("Cassandra connection credentialsSecret set to", "name", cassandraConnectionSecretName, "usernameKey", cassandraConnectionSecretUsernameKey, "passwordKey", cassandraConnectionSecretPasswordKey)
		// Set the values in the new cassandra connection
		newCassandra.Object["credentialsSecret"] = map[string]interface{}{
			"name":        cassandraConnectionSecretName,
			"usernameKey": cassandraConnectionSecretUsernameKey,
			"passwordKey": cassandraConnectionSecretPasswordKey,
		}

		return newCassandra
	}

	// spec.cassandra.connection conversion
	if conn := convertCassandraConnectionSpec(oldSpec); conn != nil && len(conn.Object) > 0 {
		err = unstructured.SetNestedField(newCassandra.Object, conn.Object, "connection")
		if err != nil {
			slog.Error("error setting cassandra connection spec", "err", err)
		}
	}

	// Always return the assembled Cassandra subresource (may be empty if missing in source)
	return newCassandra
}

func convertRabbitMQConnectionSpec(oldSpec *unstructured.Unstructured) (newConnection *unstructured.Unstructured) {
	slog.Info("Converting RabbitMQ connection spec")
	newConnection = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldConnection, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "rabbitmq", "connection")
	if err != nil {
		slog.Error("error retrieving rabbitmq connection spec", "err", err)
	}

	if !found || oldConnection == nil {
		slog.Error("spec.rabbitmq.connection section is missing or empty in the input CR. Resulting CR will have no rabbitmq connection spec resulting in a invalid Astarte CR.")
	}

	// The following fields are deep copied from the old connection to the new connection. No changes here
	dc := []string{"host", "port", "sslConfiguration", "virtualHost"}
	for _, f := range dc {
		sourcePath := []string{"spec", "rabbitmq", "connection", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newConnection, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Warn("spec.rabbitmq.connection.username is no longer supported in v2alpha1 and will be ignored. You will need to set it in the credentialsSecret.")
	slog.Warn("spec.rabbitmq.connection.password is no longer supported in v2alpha1 and will be ignored. You will need to set it in the credentialsSecret.")

	// spec.rabbitmq.connection.secret -> spec.rabbitmq.connection.credentialsSecret
	sourcePath := []string{"spec", "rabbitmq", "connection", "secret"}
	destPath := []string{"credentialsSecret"}
	err = migrationutils.CopyIfExists(oldSpec, newConnection, sourcePath, destPath)
	if err != nil {
		slog.Error("error copying secret to credentialsSecret", "err", err)
	}

	// Ask for port if not set
	if port, found, _ := unstructured.NestedInt64(newConnection.Object, "port"); !found || port == 0 {
		slog.Warn("rabbitmq.connection.port not set, please provide it:")
		_, err = fmt.Scanln(&port)

		if err != nil {
			slog.Error("error reading rabbitmq connection port from input", "err", err)
		}

		newConnection.Object["port"] = int64(port)
		slog.Info(fmt.Sprintf("rabbitmq.connection.port set to %d", port))
	}

	// Ask for host if not set
	if host, found, _ := unstructured.NestedString(newConnection.Object, "host"); !found || host == "" {
		slog.Warn("rabbitmq.connection.host not set, please provide it:")
		_, err := fmt.Scanln(&host)

		if err != nil {
			slog.Error("error reading rabbitmq connection host from input", "err", err)
		}

		newConnection.Object["host"] = host
		slog.Info(fmt.Sprintf("rabbitmq.connection.host set to %s", host))
	}

	slog.Info("RabbitMQ connection spec conversion completed")
	return newConnection
}

// convertRabbitMQSpec converts the spec.rabbitmq section from v1alpha3 to v2alpha1
func convertRabbitMQSpec(oldSpec *unstructured.Unstructured) (newRabbitMQ *unstructured.Unstructured) {
	slog.Info("Converting RabbitMQ spec")
	newRabbitMQ = &unstructured.Unstructured{Object: map[string]interface{}{}}

	oldRabbitMQ, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "rabbitmq")

	if err != nil {
		slog.Error("error retrieving rabbitmq spec", "err", err)
	}

	if !found || oldRabbitMQ == nil {
		slog.Error("spec.rabbitmq section is missing or empty in the input CR. Resulting CR will have no rabbitmq spec resulting in a invalid Astarte CR.")
		return newRabbitMQ
	}

	// The following fields are deep copied from the old rabbitmq to the new rabbitmq. No changes here
	dc := []string{"dataQueuesPrefix", "eventsExchangeName"}
	for _, f := range dc {
		sourcePath := []string{"spec", "rabbitmq", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newRabbitMQ, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("The following fields are no longer supported and will be ignored if present in the source CR: rabbitmq.deploy, rabbitmq.replicas, rabbitmq.image, rabbitmq.version, rabbitmq.storage, rabbitmq.resources, rabbitmq.additionalPlugins, rabbitmq.antiAffinity, rabbitmq.customAffinity.")

	// spec.rabbitmq.deploy is true, ask user for rabbitmq.connection.credentialsSecret details
	deploy, foundDepoy, errDepoy := unstructured.NestedBool(oldSpec.Object, "spec", "rabbitmq", "deploy")
	if deploy && foundDepoy && errDepoy == nil {
		slog.Error("spec.rabbitmq.deploy is set to true. With the new CR, you will need to deploy RabbitMQ yourself and configure rabbitmq.connection accordingly. Not doing so will result in a broken Astarte deployment.")

		rabbitmqConnectionSecretName := ""
		rabbitmqConnectionSecretUsernameKey := ""
		rabbitmqConnectionSecretPasswordKey := ""

		// Since the deploy was ture, we ask user for pointers to the new connection
		fmt.Print("new rabbitmq.connection.credentialsSecret.name: ")
		_, err := fmt.Scanln(&rabbitmqConnectionSecretName)
		if err != nil {
			slog.Error("error reading rabbitmq connection credentialsSecret name from input", "err", err)
		}

		fmt.Print("new rabbitmq.connection.credentialsSecret.usernameKey: ")
		_, err = fmt.Scanln(&rabbitmqConnectionSecretUsernameKey)
		if err != nil {
			slog.Error("error reading rabbitmq connection credentialsSecret usernameKey from input", "err", err)
		}

		fmt.Print("new rabbitmq.connection.credentialsSecret.passwordKey: ")
		_, err = fmt.Scanln(&rabbitmqConnectionSecretPasswordKey)
		if err != nil {
			slog.Error("error reading rabbitmq connection credentialsSecret passwordKey from input", "err", err)
		}

		slog.Info("RabbitMQ connection credentialsSecret set to", "name", rabbitmqConnectionSecretName, "usernameKey", rabbitmqConnectionSecretUsernameKey, "passwordKey", rabbitmqConnectionSecretPasswordKey)
		// Set the values in the new rabbitmq connection
		newRabbitMQ.Object["connection"] = map[string]interface{}{
			"credentialsSecret": map[string]interface{}{
				"name":        rabbitmqConnectionSecretName,
				"usernameKey": rabbitmqConnectionSecretUsernameKey,
				"passwordKey": rabbitmqConnectionSecretPasswordKey,
			},
		}

		return newRabbitMQ
	}

	// spec.rabbitmq.connection conversion
	if conn := convertRabbitMQConnectionSpec(oldSpec); conn != nil && len(conn.Object) > 0 {
		if unstructured.SetNestedField(newRabbitMQ.Object, conn.Object, "connection") != nil {
			slog.Error("error setting rabbitmq connection spec", "err", err)
		}
	}

	slog.Info("RabbitMQ spec conversion completed")
	return newRabbitMQ
}

// convertVernemqSpec converts the spec.vernemq section from v1alpha3 to v2alpha1
func convertVernemqSpec(oldSpec *unstructured.Unstructured) (newVernemq *unstructured.Unstructured) {
	slog.Info("Converting VerneMQ spec")
	newVernemq = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldVernemq, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "vernemq")
	if err != nil {
		slog.Error("error retrieving vernemq spec", "err", err)
	}
	if !found || oldVernemq == nil {
		slog.Error("spec.vernemq section is missing or empty in the input CR. Resulting CR will have no vernemq spec resulting in a invalid Astarte CR.")
		return newVernemq
	}

	// The following fields are deep copied from the old vernemq to the new vernemq. No changes here
	dc := getAstarteGenericClusteredResourceFields()
	// Add vernemq-specific fields
	dc2 := []string{
		"host",
		"port",
		"storage",
		"caSecret",
		"deviceHeartbeatSeconds",
		"maxOfflineMessages",
		"persistentClientExpiration",
		"mirrorQueue",
		"sslListener",
		"sslListenerCertSecretName",
	}

	dc = append(dc, dc2...)

	for _, f := range dc {
		sourcePath := []string{"spec", "vernemq", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newVernemq, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	// Default port if not set
	if port, found, _ := unstructured.NestedInt64(newVernemq.Object, "port"); !found || port == 0 {
		newVernemq.Object["port"] = int64(1883)
		slog.Warn("vernemq.port not set, defaulting to 1883. Ensure this is correct.")
	}

	slog.Info("VerneMQ spec conversion completed")
	return newVernemq
}

func convertCsrRootCaSpec(oldSpec *unstructured.Unstructured) (newCsrRootCa *unstructured.Unstructured) {
	slog.Info("Converting CFSSL csrRootCa spec")
	newCsrRootCa = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldCsrRootCa, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cfssl", "csrRootCa")
	if err != nil {
		slog.Error("error retrieving cfssl csrRootCa spec", "err", err)
	}
	if !found || oldCsrRootCa == nil {
		slog.Warn("spec.cfssl.csrRootCa section is missing or empty in the input CR. Resulting CR will have no cfssl csrRootCa spec.")
		return newCsrRootCa
	}

	// The following fields are deep copied from the old csrRootCa to the new csrRootCa. No changes here
	dc := []string{"CN", "key", "names"}
	for _, f := range dc {
		sourcePath := []string{"spec", "cfssl", "csrRootCa", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newCsrRootCa, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	// spec.cfssl.csrRootCa.ca.expiry -> spec.cfssl.csrRootCa.expiry
	sourcePath := []string{"spec", "cfssl", "csrRootCa", "ca", "expiry"}
	destPath := []string{"expiry"}
	err = migrationutils.CopyIfExists(oldSpec, newCsrRootCa, sourcePath, destPath)
	if err != nil {
		slog.Error("error copying ca.expiry to expiry", "err", err)
	}

	slog.Info("Converting CFSSL csrRootCa spec")
	return newCsrRootCa
}

func convertCaRootConfig(oldSpec *unstructured.Unstructured) (newCaRootConfig *unstructured.Unstructured) {
	slog.Info("Converting CFSSL caRootConfig spec")
	newCaRootConfig = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldCaRootConfig, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cfssl", "caRootConfig")
	if err != nil {
		slog.Error("error retrieving cfssl caRootConfig spec", "err", err)
	}
	if !found || oldCaRootConfig == nil {
		slog.Warn("spec.cfssl.caRootConfig section is missing or empty in the input CR. Resulting CR will have no cfssl caRootConfig spec.")
		return newCaRootConfig
	}
	// The object previously at `signing.default` is now directly at `signingDefault`
	sourcePath := []string{"spec", "cfssl", "caRootConfig", "signing", "default"}
	destPath := []string{"signingDefault"}
	err = migrationutils.CopyIfExists(oldSpec, newCaRootConfig, sourcePath, destPath)
	if err != nil {
		slog.Error("error copying signing.default to signingDefault", "err", err)
	}
	slog.Info("caRootConfig conversion completed")
	return newCaRootConfig
}

// convertCfsslSpec converts the spec.cfssl section from v1alpha3 to v2alpha1
func convertCfsslSpec(oldSpec *unstructured.Unstructured) (newCfssl *unstructured.Unstructured) {
	slog.Info("Converting CFSSL spec")
	newCfssl = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldCfssl, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "cfssl")
	if err != nil {
		slog.Error("error retrieving cfssl spec", "err", err)
	}
	if !found || oldCfssl == nil {
		slog.Warn("spec.cfssl section is missing or empty in the input CR. Resulting CR will have no cfssl spec.")
		return newCfssl
	}

	// The following fields are deep copied from the old cfssl to the new cfssl. No changes here

	dc := []string{
		"deploy",
		"url",
		"caExpiry",
		"caSecret",
		"certificateExpiry",
		"dbConfig",
		"resources",
		"version",
		"image",
		"storage",
		"podLabels",
		"priorityClass",
	}

	for _, f := range dc {
		sourcePath := []string{"spec", "cfssl", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newCfssl, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	// spec.cfssl.csrRootCa conversion
	if csr := convertCsrRootCaSpec(oldSpec); csr != nil && len(csr.Object) > 0 {
		if unstructured.SetNestedField(newCfssl.Object, csr.Object, "csrRootCa") != nil {
			slog.Error("error setting cfssl csrRootCa spec", "err", err)
		}
	}

	// spec.cfssl.caRootConfig conversion
	if caRoot := convertCaRootConfig(oldSpec); caRoot != nil && len(caRoot.Object) > 0 {
		if unstructured.SetNestedField(newCfssl.Object, caRoot.Object, "caRootConfig") != nil {
			slog.Error("error setting cfssl caRootConfig spec", "err", err)
		}
	}

	slog.Info("CFSSL spec conversion completed")
	return newCfssl
}

// convertComponentsSpec converts the spec.components section from v1alpha3 to v2alpha1
func convertComponentsSpec(oldSpec *unstructured.Unstructured) (newComponents *unstructured.Unstructured) {
	slog.Info("Converting Components spec")
	newComponents = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldComponents, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components")
	if err != nil {
		slog.Error("error retrieving components spec", "err", err)
	}
	if !found || oldComponents == nil {
		slog.Warn("spec.components section is missing or empty in the input CR. Resulting CR will have no components spec.")
		return newComponents
	}

	if dup := convertDataUpdaterPlantSpec(oldSpec); dup != nil && len(dup.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, dup.Object, "dataUpdaterPlant") != nil {
			slog.Error("error setting dataUpdaterPlant spec", "err", err)
		}
	}

	if te := convertTriggerEngineSpec(oldSpec); te != nil && len(te.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, te.Object, "triggerEngine") != nil {
			slog.Error("error setting triggerEngine spec", "err", err)
		}
	}

	if ae := convertAppengineApiSpec(oldSpec); ae != nil && len(ae.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, ae.Object, "appengineApi") != nil {
			slog.Error("error setting appengineApi spec", "err", err)
		}
	}

	if dash := convertDashboardSpec(oldSpec); dash != nil && len(dash.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, dash.Object, "dashboard") != nil {
			slog.Error("error setting dashboard spec", "err", err)
		}
	}

	if flow := convertFlowSpec(oldSpec); flow != nil && len(flow.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, flow.Object, "flow") != nil {
			slog.Error("error setting flow spec", "err", err)
		}
	}

	// Still to convert:
	// Pairing API and Backend
	// Housekeeping API and Backend
	// Realm Management API and Backend
	// For these services, we need to merge the API and Backend specs into a single spec for each service
	// resources (memory/cpu) limits and requests are summed together

	if pairing := convertPairingSpec(oldSpec); pairing != nil && len(pairing.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, pairing.Object, "pairing") != nil {
			slog.Error("error setting pairing spec", "err", err)
		}
	}

	if housekeeping := convertAstarteGenericComponentSpec(oldSpec, "housekeeping"); housekeeping != nil && len(housekeeping.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, housekeeping.Object, "housekeeping") != nil {
			slog.Error("error setting housekeeping spec", "err", err)
		}
	}

	if realm := convertAstarteGenericComponentSpec(oldSpec, "realmManagement"); realm != nil && len(realm.Object) > 0 {
		if unstructured.SetNestedField(newComponents.Object, realm.Object, "realmManagement") != nil {
			slog.Error("error setting realmManagement spec", "err", err)
		}
	}

	slog.Info("Components spec conversion completed")
	return newComponents
}

// convertPairingSpec converts the spec.components.pairingApi and spec.components.pairingBackend sections from v1alpha3 to v2alpha1
func convertPairingSpec(oldSpec *unstructured.Unstructured) (newPairing *unstructured.Unstructured) {
	slog.Info("Converting Pairing spec: merging API and Backend specs")
	newPairing = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldPairing, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "pairing")
	if err != nil {
		slog.Error("error retrieving pairing spec", "err", err)
	}
	if !found || oldPairing == nil {
		slog.Warn("spec.components.pairing section is missing or empty in the input CR. Resulting CR will have no pairing spec.")
		return newPairing
	}

	// Use convertAstarteGenericComponentSpec
	if newPairing = convertAstarteGenericComponentSpec(oldSpec, "pairing"); newPairing == nil {
		slog.Error("error converting pairing spec")
		return nil
	}

	slog.Info("Pairing spec conversion completed")
	return newPairing
}

// convertDataUpdaterPlantSpec converts the spec.components.dataUpdaterPlant section from v1alpha3 to v2alpha1
// DUP is basically unchanged between v1alpha3 and v2alpha1
func convertDataUpdaterPlantSpec(oldSpec *unstructured.Unstructured) (newDataUpdaterPlant *unstructured.Unstructured) {
	slog.Info("Converting DataUpdaterPlant spec")
	newDataUpdaterPlant = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldDataUpdaterPlant, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "dataUpdaterPlant")
	if err != nil {
		slog.Error("error retrieving dataUpdaterPlant spec", "err", err)
	}
	if !found || oldDataUpdaterPlant == nil {
		slog.Warn("spec.components.dataUpdaterPlant section is missing or empty in the input CR. Resulting CR will have no dataUpdaterPlant spec.")
		return newDataUpdaterPlant
	}

	// The following fields are deep copied from the old dataUpdaterPlant to the new dataUpdaterPlant. No changes here
	dc1 := []string{
		"prefetchCount",
		"dataQueueCount",
	}

	dc2 := getAstarteGenericClusteredResourceFields()

	dc := append(dc1, dc2...)
	for _, f := range dc {
		sourcePath := []string{"spec", "components", "dataUpdaterPlant", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newDataUpdaterPlant, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("DataUpdaterPlant spec conversion completed")
	return newDataUpdaterPlant
}

// mergeAdditionalEnv merges api and backend additionalEnv for a given component.
// Returns the merged unstructured value and true if any env was found.
func mergeAdditionalEnv(oldSpec *unstructured.Unstructured, componentName string) (interface{}, bool) {
	aEnv, apiEnvFound, apiEnvErr := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "api", "additionalEnv")
	if apiEnvErr != nil {
		slog.Error("error retrieving "+componentName+" api additionalEnv", "err", apiEnvErr)
	}

	bEnv, backendEnvFound, backendEnvErr := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "backend", "additionalEnv")
	if backendEnvErr != nil {
		slog.Error("error retrieving "+componentName+" backend additionalEnv", "err", backendEnvErr)
	}

	if apiEnvFound && !backendEnvFound {
		return aEnv, true
	}

	if !apiEnvFound && backendEnvFound {
		return bEnv, true
	}

	if apiEnvFound && backendEnvFound {
		aEnvList, convErrA := migrationutils.UnstructuredToEnvVarList(aEnv)
		if convErrA != nil {
			slog.Error("error converting api additionalEnv", "err", convErrA)
		}

		bEnvList, convErrB := migrationutils.UnstructuredToEnvVarList(bEnv)
		if convErrB != nil {
			slog.Error("error converting backend additionalEnv", "err", convErrB)
		}

		if mergedEnv := migrationutils.MergeAdditionalEnv(aEnvList, bEnvList); mergedEnv != nil {
			return migrationutils.EnvVarListToUnstructured(mergedEnv), true
		}
	}

	return nil, false
}

// mergeResources merges api and backend resources for a given component.
// Returns the merged unstructured value and true if any resources were found.
func mergeResources(oldSpec *unstructured.Unstructured, componentName string) (interface{}, bool, error) {
	aRes, apiFound, apiErr := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "api", "resources")
	if apiErr != nil {
		slog.Error("error retrieving "+componentName+" api resources", "err", apiErr)
	}

	bRes, backendFound, backendErr := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "backend", "resources")
	if backendErr != nil {
		slog.Error("error retrieving "+componentName+" backend resources", "err", backendErr)
	}

	if apiFound && !backendFound {
		aRR, errA := migrationutils.UnstructuredToResourceRequirements(aRes)
		if errA != nil {
			slog.Error("error converting api resources", "err", errA)
			return nil, true, fmt.Errorf("error converting api resources for %s: %w", componentName, errA)
		}
		converted, err := migrationutils.ResourceRequirementsToUnstructured(aRR)
		if err != nil {
			slog.Error("error converting api resources to unstructured", "err", err)
			return nil, true, fmt.Errorf("error converting api resources to unstructured for %s: %w", componentName, err)
		}
		return converted, true, nil
	}

	if !apiFound && backendFound {
		bRR, errB := migrationutils.UnstructuredToResourceRequirements(bRes)
		if errB != nil {
			slog.Error("error converting backend resources", "err", errB)
			return nil, true, fmt.Errorf("error converting backend resources for %s: %w", componentName, errB)
		}
		converted, err := migrationutils.ResourceRequirementsToUnstructured(bRR)
		if err != nil {
			slog.Error("error converting backend resources to unstructured", "err", err)
			return nil, true, fmt.Errorf("error converting backend resources to unstructured for %s: %w", componentName, err)
		}
		return converted, true, nil
	}

	if apiFound && backendFound {
		aRR, errA := migrationutils.UnstructuredToResourceRequirements(aRes)
		if errA != nil {
			slog.Error("error converting api resources", "err", errA)
			// Decide if this should be a fatal error or continue with partial data
		}
		bRR, errB := migrationutils.UnstructuredToResourceRequirements(bRes)
		if errB != nil {
			slog.Error("error converting backend resources", "err", errB)
			// Decide if this should be a fatal error or continue with partial data
		}
		mergedRes := migrationutils.SumResourceRequirements(aRR, bRR)
		converted, err := migrationutils.ResourceRequirementsToUnstructured(mergedRes)
		if err != nil {
			slog.Error("error converting merged resources to unstructured", "err", err)
			return nil, true, fmt.Errorf("error converting merged resources to unstructured for %s: %w", componentName, err)
		}
		return converted, true, nil
	}

	return nil, false, nil
}

// convertAstarteGenericComponentSpec merges the API and Backend specs of a generic Astarte component (housekeeping, realmManagement, pairing)
func convertAstarteGenericComponentSpec(oldSpec *unstructured.Unstructured, componentName string) (newComponent *unstructured.Unstructured) {
	slog.Info("Converting " + componentName + " spec")
	newComponent = &unstructured.Unstructured{Object: map[string]interface{}{}}
	if componentName != "housekeeping" && componentName != "realmManagement" && componentName != "pairing" {
		slog.Error("convertAstarteGenericComponentSpec called with unsupported component name: " + componentName)
		return nil
	}

	// Each generic component has two sub-sections in v1alpha3: API and Backend (api and backend)
	// API contains fields of AstarteGenericAPISpec (see getAstarteGenericAPISpecFields)
	// Backend contains fields of AstarteGenericClusteredResource (see getAstarteGenericClusteredResourceFields)

	// Fetch api and backend subsections from old spec
	oldAPI, apiFound, errAPI := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "api")
	if errAPI != nil {
		slog.Error("error retrieving "+componentName+" api spec", "err", errAPI)
	}

	// Fetch backend subsection from old spec
	oldBackend, backendFound, errBackend := unstructured.NestedFieldCopy(oldSpec.Object, "spec", "components", componentName, "backend")
	if errBackend != nil {
		slog.Error("error retrieving "+componentName+" backend spec", "err", errBackend)
	}

	// If neither api nor backend found, return empty newComponent
	if (!apiFound || oldAPI == nil) && (!backendFound || oldBackend == nil) {
		slog.Warn("spec.components." + componentName + " section has neither api nor backend in the source CR")
		return newComponent
	}

	// If only api found, copy it entirely to newComponent
	if apiFound && (!backendFound || oldBackend == nil) {
		slog.Info(componentName + " has only api spec, copying it entirely")
		if apiMap, ok := oldAPI.(map[string]interface{}); ok {
			// Deep-copy to avoid sharing references with source
			copied := make(map[string]interface{}, len(apiMap))
			for k, v := range apiMap {
				copied[k] = v
			}
			newComponent.Object = copied
		} else {
			slog.Error("error converting " + componentName + " api spec to map")
		}
		slog.Info(componentName + " spec conversion completed")
		return newComponent
	}

	// If only backend found, copy it entirely to newComponent
	if backendFound && (!apiFound || oldAPI == nil) {
		slog.Info(componentName + " has only backend spec, copying it entirely")
		if backendMap, ok := oldBackend.(map[string]interface{}); ok {
			// Deep-copy to avoid sharing references with source
			copied := make(map[string]interface{}, len(backendMap))
			for k, v := range backendMap {
				copied[k] = v
			}
			newComponent.Object = copied
		} else {
			slog.Error("error converting " + componentName + " backend spec to map")
		}
		slog.Info(componentName + " spec conversion completed")
		return newComponent
	}

	// Both api and backend found. Merge them accordingly

	// For most fields, check if they exist in either API or Backend and copy them to the top-level of the new component spec,
	// if they exist in both, the Backend value takes precedence. Exceptions:
	// - RESOURCES (memory/cpu) limits and requests are summed together if present in both API and Backend
	// - Environment variables are merged, with Backend variables taking precedence in case of conflicts

	// First handle resources via helper
	mergedRes, found, err := mergeResources(oldSpec, componentName)
	if err != nil {
		slog.Error("error merging resources for "+componentName, "err", err)
	}
	if found {
		newComponent.Object["resources"] = mergedRes
	}

	// Handle additionalEnv specially by merging arrays with backend precedence.
	if mergedEnv, found := mergeAdditionalEnv(oldSpec, componentName); found {
		newComponent.Object["additionalEnv"] = mergedEnv
	}

	// Build the union of candidate fields from API and Backend specs.
	// These are the only fields we will consider copying.
	unionFields := map[string]struct{}{}
	for _, f := range getAstarteGenericAPISpecFields() {
		unionFields[f] = struct{}{}
	}
	for _, f := range getAstarteGenericClusteredResourceFields() {
		unionFields[f] = struct{}{}
	}

	var apiMap map[string]interface{}
	var backendMap map[string]interface{}

	if oldAPIMap, ok := oldAPI.(map[string]interface{}); ok {
		apiMap = oldAPIMap
	} else {
		slog.Error("error converting " + componentName + " api spec to map")
	}

	if oldBackendMap, ok := oldBackend.(map[string]interface{}); ok {
		backendMap = oldBackendMap
	} else {
		slog.Error("error converting " + componentName + " backend spec to map")
	}

	// For the remaining fields, copy with backend precedence, falling back to api
	for f := range unionFields {
		if f == "resources" || f == "additionalEnv" {
			continue // already handled
		}
		var val interface{}
		if backendMap != nil {
			if v, ok := backendMap[f]; ok && v != nil {
				val = v
			}
		}
		if val == nil && apiMap != nil {
			if v, ok := apiMap[f]; ok && v != nil {
				val = v
			}
		}
		if val != nil {
			newComponent.Object[f] = val
		}
	}

	slog.Info(componentName + " spec conversion completed")

	return newComponent
}

// convertAppengineApiSpec converts the spec.components.appengineApi section from v1alpha3 to v2alpha1
// AppEngine API is basically unchanged between v1alpha3 and v2alpha1
func convertAppengineApiSpec(oldSpec *unstructured.Unstructured) (newAppengineApi *unstructured.Unstructured) {
	slog.Info("Converting AppEngine API spec")
	newAppengineApi = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldAppengineApi, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "appengineApi")
	if err != nil {
		slog.Error("error retrieving appengineApi spec", "err", err)
	}
	if !found || oldAppengineApi == nil {
		slog.Warn("spec.components.appengineApi section is missing or empty in the input CR. Resulting CR will have no appengineApi spec.")
		return newAppengineApi
	}

	// The following fields are deep copied from the old appengineApi to the new appengineApi. No changes here
	dc1 := []string{
		"maxResultsLimit",
		"roomEventsQueueName",
		"roomEventsExchangeName",
	}

	dc2 := getAstarteGenericAPISpecFields()

	dc := append(dc1, dc2...)
	for _, f := range dc {
		sourcePath := []string{"spec", "components", "appengineApi", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newAppengineApi, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("AppEngine API spec conversion completed")
	return newAppengineApi
}

// convertDashboardSpec converts the spec.components.dashboard section from v1alpha3 to v2alpha1
// Dashboard is basically unchanged between v1alpha3 and v2alpha1
func convertDashboardSpec(oldSpec *unstructured.Unstructured) (newDashboardApi *unstructured.Unstructured) {
	slog.Info("Converting Dashboard spec")
	newDashboardApi = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldDashboardApi, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "dashboard")
	if err != nil {
		slog.Error("error retrieving dashboard spec", "err", err)
	}
	if !found || oldDashboardApi == nil {
		slog.Warn("spec.components.dashboard section is missing or empty in the input CR. Resulting CR will have no dashboard spec.")
		return newDashboardApi
	}

	// The following fields are deep copied from the old dashboard to the new dashboard. No changes here
	dc1 := []string{
		"realmManagementApiUrl",
		"appEngineApiUrl",
		"pairingApiUrl",
		"flowApiUrl",
		"defaultRealm",
		"defaultAuth",
		"auth",
	}

	dc2 := getAstarteGenericAPISpecFields()

	dc := append(dc1, dc2...)
	for _, f := range dc {
		sourcePath := []string{"spec", "components", "dashboard", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newDashboardApi, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("Dashboard spec conversion completed")
	return newDashboardApi
}

// convertFlowSpec converts the spec.components.flow section from v1alpha3 to v2alpha1
// Flow is basically unchanged between v1alpha3 and v2alpha1
func convertFlowSpec(oldSpec *unstructured.Unstructured) (newFlow *unstructured.Unstructured) {
	slog.Info("Converting Flow spec")
	newFlow = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldFlow, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "flow")
	if err != nil {
		slog.Error("error retrieving flow spec", "err", err)
	}
	if !found || oldFlow == nil {
		slog.Warn("spec.components.flow section is missing or empty in the input CR. Resulting CR will have no flow spec.")
		return newFlow
	}

	// The following fields are deep copied from the old flow to the new flow. No changes here
	dc := getAstarteGenericAPISpecFields()

	for _, f := range dc {
		sourcePath := []string{"spec", "components", "flow", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newFlow, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("Flow spec conversion completed")
	return newFlow
}

// convertTriggerEngineSpec converts the spec.components.triggerEngine section from v1alpha3 to v2alpha1
// TE is basically unchanged between v1alpha3 and v2alpha1
func convertTriggerEngineSpec(oldSpec *unstructured.Unstructured) (newTriggerEngine *unstructured.Unstructured) {
	slog.Info("Converting TriggerEngine spec")
	newTriggerEngine = &unstructured.Unstructured{Object: map[string]interface{}{}}
	oldTriggerEngine, found, err := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "components", "triggerEngine")
	if err != nil {
		slog.Error("error retrieving triggerEngine spec", "err", err)
	}
	if !found || oldTriggerEngine == nil {
		slog.Warn("spec.components.triggerEngine section is missing or empty in the input CR. Resulting CR will have no triggerEngine spec.")
		return newTriggerEngine
	}

	// The following fields are deep copied from the old triggerEngine to the new triggerEngine. No changes here
	dc1 := getAstarteGenericClusteredResourceFields()
	dc2 := []string{
		"eventsQueueName",
		"eventsRoutingKey",
	}
	dc := append(dc1, dc2...)
	for _, f := range dc {
		sourcePath := []string{"spec", "components", "triggerEngine", f}
		destPath := []string{f}
		err = migrationutils.CopyIfExists(oldSpec, newTriggerEngine, sourcePath, destPath)
		if err != nil {
			slog.Error("error copying field", "field", f, "err", err)
		}
	}

	slog.Info("TriggerEngine spec conversion completed")
	return newTriggerEngine
}

// convertSpec converts the spec section from v1alpha3 to v2alpha1
func convertSpec(oldSpec *unstructured.Unstructured) (newSpec *unstructured.Unstructured) {

	// Initialize the destination Unstructured with a non-nil Object map
	newSpec = &unstructured.Unstructured{Object: map[string]interface{}{}}

	// Check if spec is nil, if so, return nil
	if oldSpec.Object == nil {
		slog.Error("spec section is missing or empty in the input CR. Resulting CR will have no spec.")
		return nil
	}

	// The following fields are deep copied from the old spec to the new spec. No changes here
	dc := []string{"api", "version", "imagePullPolicy", "imagePullSecrets", "distributionChannel", "deploymentStrategy", "features", "storageClassName", "astarteInstanceID", "manualMaintenanceMode"}
	for _, f := range dc {
		sourcePath := []string{"spec", f}
		// We are building the contents of spec here, so destination must be the root of newSpec
		destPath := []string{f}
		_ = migrationutils.CopyIfExists(oldSpec, newSpec, sourcePath, destPath)
	}

	// Check for and warn about removed fields
	if rbac, found, _ := unstructured.NestedFieldNoCopy(oldSpec.Object, "spec", "rbac"); found && rbac != nil {
		slog.Warn("spec.rbac field is no longer supported and will be ignored. RBAC is now always managed by the operator.")
	}

	// Cassandra: spec.cassandra conversion
	if cass := convertCassandraSpec(oldSpec); cass != nil && len(cass.Object) > 0 {
		if unstructured.SetNestedField(newSpec.Object, cass.Object, "cassandra") != nil {
			slog.Error("error setting cassandra spec")
		}
	}

	// RabbitMQ: spec.rabbitmq conversion
	if rmq := convertRabbitMQSpec(oldSpec); rmq != nil && len(rmq.Object) > 0 {
		if unstructured.SetNestedField(newSpec.Object, rmq.Object, "rabbitmq") != nil {
			slog.Error("error setting rabbitmq spec")
		}
	}

	// VerneMQ: spec.vernemq conversion
	if vmq := convertVernemqSpec(oldSpec); vmq != nil && len(vmq.Object) > 0 {
		if unstructured.SetNestedField(newSpec.Object, vmq.Object, "vernemq") != nil {
			slog.Error("error setting vernemq spec")
		}
	}

	// CFSSL: spec.cfssl conversion
	if cfssl := convertCfsslSpec(oldSpec); cfssl != nil && len(cfssl.Object) > 0 {
		if unstructured.SetNestedField(newSpec.Object, cfssl.Object, "cfssl") != nil {
			slog.Error("error setting cfssl spec")
		}
	}

	// Components: spec.components conversion
	if comps := convertComponentsSpec(oldSpec); comps != nil && len(comps.Object) > 0 {
		if unstructured.SetNestedField(newSpec.Object, comps.Object, "components") != nil {
			slog.Error("error setting components spec")
		}
	}

	return newSpec
}

// v1alpha3toV2alpha1 converts a v1alpha3 Astarte CR to a v2alpha1 CR
func V1alpha3toV2alpha1(oldCr *unstructured.Unstructured) (*unstructured.Unstructured, error) {

	// Disclamer: this code is very verbose and not very elegant, i know. The goal here
	// is to have a very explicit and clear procedural approach to the conversion, so that
	// we can easily spot what is being converted, what is being ignored and what needs
	// user intervention. This is a one-time use tool, so maintainability and elegance
	// are not a priority here. I swear, I usually write much better code.

	// Initialize the destination Unstructured with a non-nil Object map
	newCr := &unstructured.Unstructured{Object: map[string]interface{}{}}

	// If not v1alpha3, return error
	apiVersion, found, err := unstructured.NestedString(oldCr.Object, "apiVersion")
	if err != nil || !found || apiVersion != "api.astarte-platform.org/v1alpha3" {
		return nil, fmt.Errorf("error: input CR is not v1alpha3. Cannot convert to v2alpha1")
	}

	// Update API version
	if unstructured.SetNestedField(newCr.Object, "api.astarte-platform.org/v2alpha1", "apiVersion") != nil {
		return nil, fmt.Errorf("error setting apiVersion in new CR: %v", err)
	}

	// Copy kind
	kind, found, err := unstructured.NestedString(oldCr.Object, "kind")
	if err != nil || !found || kind != "Astarte" {
		return nil, fmt.Errorf("error: input CR kind is not Astarte. Cannot convert to v2alpha1")
	}

	if unstructured.SetNestedField(newCr.Object, kind, "kind") != nil {
		return nil, fmt.Errorf("error setting kind in new CR: %v", err)
	}

	// Copy name and namespace from metadata
	name, found, err := unstructured.NestedString(oldCr.Object, "metadata", "name")
	if err != nil || !found {
		return nil, fmt.Errorf("error: input CR metadata.name not found. Cannot convert to v2alpha1")
	}

	if unstructured.SetNestedField(newCr.Object, name, "metadata", "name") != nil {
		return nil, fmt.Errorf("error setting metadata.name in new CR: %v", err)
	}

	// Copy namespace only if found
	namespace, found, err := unstructured.NestedString(oldCr.Object, "metadata", "namespace")
	if err != nil {
		return nil, fmt.Errorf("error: input CR metadata.namespace. Cannot convert to v2alpha1")
	}

	if found {
		if err := unstructured.SetNestedField(newCr.Object, namespace, "metadata", "namespace"); err != nil {
			return nil, fmt.Errorf("error setting metadata.namespace in new CR: %v", err)
		}
	}

	// Assign converted spec to newCr
	convertedSpec := convertSpec(oldCr)
	if convertedSpec != nil {
		// Set the underlying map (convertedSpec.Object) as the spec
		if err = unstructured.SetNestedField(newCr.Object, convertedSpec.Object, "spec"); err != nil {
			return nil, fmt.Errorf("error setting spec in new CR: %v", err)
		}
	} else {
		slog.Warn("spec section is missing or empty in the input CR. Resulting CR will have no spec.")
	}

	return newCr, nil
}

// getAstarteGenericClusteredResourceFields returns a list of fields that are common to all Astarte Backend components
func getAstarteGenericClusteredResourceFields() []string {
	return []string{
		"deploy",
		"replicas",
		"antiAffinity",
		"customAffinity",
		"deploymentStrategy",
		"version",
		"image",
		"resources",
		"additionalEnv",
		"podLabels",
		"autoscaler",
		"priorityClass",
	}
}

// getAstarteGenericAPISpecFields returns a list of fields that are common to all Astarte API components
func getAstarteGenericAPISpecFields() []string {
	// Actually, all fields are common except disableAuthentication
	// which is only present in some APIs
	fields := getAstarteGenericClusteredResourceFields()
	fields = append(fields, []string{
		"disableAuthentication",
	}...)
	return fields
}
