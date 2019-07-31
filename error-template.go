package main

const errorXML = `
<?xml version="1.0"?>
<ows:ExceptionReport xmlns:ows="http://www.opengis.net/ows/1.1"
                     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
                     xsi:schemaLocation="http://www.opengis.net/ows/1.1 http://schemas.opengis.net/ows/1.1.0/owsExceptionReport.xsd"
                     version="1.0.0" xml:lang="en">
    <ows:Exception exceptionCode="{{ .ExceptionCode }}">
        <ows:ExceptionText>{{ .Message }}</ows:ExceptionText>
    </ows:Exception>
</ows:ExceptionReport>`
