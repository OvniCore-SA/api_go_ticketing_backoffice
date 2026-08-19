package entities

// Nota: la tabla de permisos se elimina en esta versión del backoffice.
// El único rol operativo es SuperAdmin, cuya autoridad es global — no hay
// necesidad de un grafo de permisos granular. Si más adelante se agregan
// roles internos con permisos parciales, reintroducir Permission + tabla
// pivot roles_has_permissions.
