# CycloneDX schema

The generated BOM follows the cycloneDX schema mentioned [here](https://cyclonedx.org/docs/1.6/json/)

# Cluster Bill of Materials (cluster BOM) Structure

This repository defines a Go data structure for a **Cluster Bill of Materials (cluster BOM)**, following standard formats for tracking software components, dependencies, and metadata.

## 📌 BOM Structure

The `BOM` struct represents the core cluster BOM, containing:

- **`bomFormat`** – Specifies the cluster BOM format.
- **`specVersion`** – Defines the specification version.
- **`serialNumber`** *(optional)* – A unique identifier for the cluster BOM.
- **`version`** – The cluster BOM version.
- **`metadata`** – Metadata related to cluster BOM generation.
- **`components`** *(optional)* – A list of software components included in the cluster BOM.

## 📝 Metadata

Metadata provides additional details about cluster BOM creation, including:

- **`timestamp`** – When the cluster BOM was generated.
- **`tools`** – List of tools that created the cluster BOM. This will be Cluster Codex.
- **`component`** – The primary software component described in the cluster BOM. For Cluster Codex this will be the Kubernetes cluster itself.

## 🔧 Components

A `Component` represents a Kubernetes object, software package or library, containing:

- **`type`** – The category of the component (e.g., Kubernetes object, library, application).
- **`name`** – The name of the component.
- **`version`** – The specific version of the component.
- **`purl`** *(optional)* – The Package URL for identification.
- **`properties`** *(optional)* – Custom key-value metadata about the component.
- **`licenses`** *(optional)* – Licensing information.
- **`hashes`** *(optional)* – Cryptographic hashes for integrity verification.

## 📂 Additional Structures

### 🛠 Tool
Identifies the tool used to generate the cluster BOM.
- **`vendor`** – The name of the vendor.
- **`name`** – The tool name.
- **`version`** – The tool version.

### 🏷 Property
A key-value pair for additional metadata.
- **`name`** – The property name.
- **`value`** – The property value.

### 📜 License
Contains licensing details.
- **`id`** *(optional)* – The license identifier.
- **`name`** – The license name.

### 🔐 Hash
Stores cryptographic hashes for component verification.
- **`alg`** – The hashing algorithm used.
- **`value`** – The computed hash value.

---

This struct provides a **structured and standardized way** to describe a cluster BOM in Go, ensuring **traceability, security, and compliance**.
