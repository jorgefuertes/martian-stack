package web

// Content Negotiation
const (
	HeaderAccept             = "Accept"              // This header informs the server about the media types (e.g., text/html, image/png) the client is willing to accept. It can include multiple types, each with an optional quality factor (q) to express preference. The web-browsers send Accept headers to indicate supported formats, allowing servers to choose the most suitable one.
	HeaderAcceptCharset      = "Accept-Charset"      // This header specifies the character encodings (e.g., UTF-8, ISO-8859-1) the client understands. It can include multiple encodings with quality factors. It is useful for serving content in the appropriate character encoding based on the client's capabilities.
	HeaderAcceptLanguage     = "Accept-Language"     // This header informs the server about the languages (e.g., en-US, fr-FR) the client prefers. It can include multiple languages with quality factors. It allows servers to deliver content in the user's preferred language, enhancing user experience.
	HeaderAcceptEncoding     = "Accept-Encoding"     // This header specifies the compression algorithms (e.g., gzip, deflate) the client can handle. It can include multiple algorithms with quality factors.
	HeaderVary               = "Vary"                // This header informs clients about response headers that are dependent on the values in specific request headers like Accept and Accept-Language. It helps clients cache responses based on negotiated content for future requests.
	HeaderContentType        = "Content-Type"        // While not directly for negotiation, this header in the server response confirms the media type of the delivered content. It validates the server's choice and helps clients process the content correctly.
	HeaderContentDisposition = "Content-Disposition" // Content-Disposition: attachment; filename="logo.png"
	HeaderContentLanguage    = "Content-Language"    // Similar to Accept-Language, this header in the server response indicates the actual language of the delivered content.
)

// Caching
const (
	HeaderCacheControl    = "Cache-Control"     // This is the most important header for caching strategies. It allows you to specify how long a resource can be cached on the client side. It also suggests if the resource can be cached by public or private caches, and how it should be revalidated.
	HeaderExpires         = "Expires"           // This header is deprecated, but it can still be used to specify an absolute expiration date for a resource.
	HeaderLastModified    = "Last-Modified"     // This header indicates the date and time that a resource was last modified. This is optional, yet it can be used by caches to validate whether a cached resource is still up-to-date.
	HeaderETag            = "ETag"              // This header is a unique identifier for a resource. It can be used by caches to validate whether a cached resource is still up-to-date, even if the Last-Modified header has not changed.
	HeaderIfModifiedSince = "If-Modified-Since" // This header can be used by a client to ask a server if a cached resource is still up-to-date. The client sends the Last-Modified header from the cached resource, and the server responds with a 304 Not Modified if the resource is still up-to-date, or with the new resource if it has been modified.
	HeaderIfNoneMatch     = "If-None-Match"     // This header can be used by a client to ask a server if a cached resource is still up-to-date. The client sends the ETag header from the cached resource, and the server responds with a 304 Not Modified if the resource is still up-to-date, or with the new resource if it has been modified.
)

// Web Security
const (
	HeaderXXSSProtection            = "X-XSS-Protection"            // This header helps prevent reflected cross-site scripting (XSS) attacks by stopping pages from loading when they detect them. However, it is not recommended to use this header, as it can create XSS vulnerabilities in otherwise safe websites.
	HeaderXFrameOptions             = "X-Frame-Options"             // This header helps prevent clickjacking attacks by stopping a page from loading within a frame or iframe.
	HeaderXContentTypeOptions       = "X-Content-Type-Options"      // This header helps prevent MIME type sniffing attacks by stopping browsers from guessing the MIME type of a resource.
	HeaderReferrerPolicy            = "Referrer-Policy"             // This header controls how much referrer information is sent with requests.
	HeaderSetCookie                 = "Set-Cookie"                  // This header is used to send cookies from the server to the user agent.
	HeaderStrictTransportSecurity   = "Strict-Transport-Security"   // (HSTS) This header tells a website to only be accessed using HTTPS.
	HeaderExpectCT                  = "Expect-CT"                   // This header is used to report Certificate Transparency (CT) requirements. It is not recommended to use this header, as it is being phased out.
	HeaderContentSecurityPolicy     = "Content-Security-Policy"     // (CSP) This header helps mitigate XSS and data injection attacks by specifying the origins of content that can be loaded on a website.
	HeaderAccessControlAllowOrigin  = "Access-Control-Allow-Origin" // (CORS) This header relaxes the Same Origin Policy (SOP) by allowing certain origins to access resources from a website.
	HeaderAccessControlAllowMethods = "Access-Control-Allow-Methods"
	HeaderAccessControlAllowHeaders = "Access-Control-Allow-Headers"
	HeaderCrossOriginOpenerPolicy   = "Cross-Origin-Opener-Policy"   // (COOP) This header helps prevent attacks like Spectre by stopping a top-level document from sharing a browsing context group with cross-origin documents.
	HeaderCrossOriginEmbedderPolicy = "Cross-Origin-Embedder-Policy" // (COEP) This header helps prevent attacks like Spectre by stopping a document from loading any cross-origin resources that don't explicitly grant the document permission.
	HeaderCrossOriginResourcePolicy = "Cross-Origin-Resource-Policy" // (CORP) This header helps prevent attacks like Spectre by controlling the set of origins that are allowed to load a resource.
	HeaderPermissionsPolicy         = "Permissions-Policy"           // This header controls which origins can use which browser features.
	HeaderFLoC                      = "FLoC"                         // (Federated Learning of Cohorts) This header allows a site to opt-out of being included in a user's list of sites for cohort calculation, which is used for interest-based advertising.
	HeaderServer                    = "Server"                       // This header describes the software used by the origin server. It is not a security header, but attackers can use the information in it to find vulnerabilities.
	HeaderXPoweredBy                = "X-Powered-By"                 // This header describes the technologies used by the web server. Attackers can use the information in it to find vulnerabilities.
	HeaderXAspNetVersion            = "X-AspNet-Version"             // This header provides information about the .NET version. It is recommended to disable sending this header.
	HeaderXAspNetMvcVersion         = "X-AspNetMvc-Version"          // This header provides information about the .NET version. It is recommended to disable sending this header.
	HeaderXDNSPrefetchControl       = "X-DNS-Prefetch-Control"       // This header controls DNS prefetching, which is a feature by which browsers proactively perform domain name resolution.
	HeaderPublicKeyPins             = "Public-Key-Pins"              // (HPKP) This header is used to associate a specific cryptographic public key with a certain web server to decrease the risk of MITM attacks with forged certificates. It is deprecated and should not be used anymore.
)

// Cross-Origin Resource Sharing

const (
	HeaderOrigin                        = "Origin"                           // (Request header) Identifies the origin of the request (originating website). This should be sent in the request from the client to the remote site.
	HeaderAccessControlAllowCredentials = "Access-Control-Allow-Credentials" // (Server Response) Specifies whether the server allows credentials (cookies, authorization headers) to be sent in cross-origin requests. This is crucial for functionalities like authenticated API calls across different domains.
	HeaderAccessControlExposeHeaders    = "Access-Control-Expose-Headers"    // (Server Response) Allows the server to explicitly list custom response headers that should be made accessible to the client-side JavaScript code, even if not included in the default CORS-accessible headers. This enables utilization of data within the client application.
	HeaderAccessControlRequestMethod    = "Access-Control-Request-Method"    // (Preflight Request) Used in preflight requests (OPTIONS method) sent by the browser before the actual request. It informs the server about the HTTP method intended for the actual request (e.g., POST, PUT).
	HeaderAccessControlRequestHeaders   = "Access-Control-Request-Headers"   // (Preflight Request) Another preflight request header informing the server about any custom request headers the browser plans to send with the actual request. This allows the server to verify compatibility and handle custom headers appropriately.
)

/*
	During a CORS request,
	the browser first sends a preflight request (OPTIONS method) with the Access-Control-Request-Method and Access-Control-Request-Headers (optional) to check if the server allows the actual request.
    the server responds with the Access-Control-Allow-Origin, Access-Control-Allow-Credentials (optional), and Access-Control-Expose-Headers (optional) headers, indicating permissions and accessible data.
    if the preflight is successful, the browser sends the actual request with the allowed method and headers.

	The Origin header (in the request) and Vary header (in the response) plays a role in CORS as well, though they are not strictly CORS-specific headers. For details specs, features, and limitations refer to MDN Docs on CORS
	HTTP Communication

	There are headers required for basic communication. HTTP offers a a bunch of headers catering to basic client/server communication.
*/

const (
	HeaderContentEncoding = "Content-Encoding" // Sent in the HTTP response. Indicates the compression method applied to the response body by the server (e.g., gzip, br). This helps the client decompress the received data accurately.
	HeaderLocation        = "Location"         // This header is present in the response. Instructs the client to redirect to a different URL, often used for error handling (e.g., 301 Moved Permanently) or load balancing purposes.
	HeaderRange           = "Range"            // Used to request a specific portion of a resource, rather than the entire content. This is particularly beneficial for large files downloads, enabling the client to download only the desired section (e.g., Range: bytes=100-200).
	HeaderContentRange    = "Content-Range"    // When responding to a range request, the server includes this header to specify the range of bytes delivered in the partial response. This helps the client understand the extent of the received data and potentially make further requests if necessary.
	HeaderAuthorization   = "Authorization"    // This is sent in HTTP requests and carries authentication credentials (e.g., username and password) to allow access to protected resources.
	HeaderAcceptCH        = "Accept-CH"        // (Request) Signals the client's support for Client Hints, a mechanism allowing the server to request specific pieces of information from the client before responding. This can improve efficiency by preemptively fetching resources the client is likely to need.
	HeaderDNT             = "DNT"              // This provides an optional signal from the user regarding their Do Not Track (DNT) preference. The header value can be either 0 (disabled) or 1 (enabled) and should be sent in the request. While not mandatory for websites to respect this header, it allows users to express their preference for limiting online tracking.
)
