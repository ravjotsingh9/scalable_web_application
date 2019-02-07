import React, { Component } from "react";
import "filepond/dist/filepond.min.css";
import { Container, Row, Col, Badge, Button } from "react-bootstrap";
import { FilePond, registerPlugin } from "react-filepond";
import FilePondPluginImagePreview from "filepond-plugin-image-preview";
import "filepond-plugin-image-preview/dist/filepond-plugin-image-preview.min.css";
registerPlugin(FilePondPluginImagePreview);

class CsvUpload extends Component {
  constructor(props) {
    super(props);
    this.state = {};
    this.render = this.render.bind(this);
    this.displayError = this.displayError.bind(this);
  }

  displayError(error) {
    console.log("Display Error------------");
    console.log(error);
  }

  render() {
    let _this = this;
    _this.api_upload_address = "http://localhost:80/uploadReport";

    return (
      <div>
        <Container>
          <Row>
            <Col>
              <FilePond
                ref={ref => (this.pond = ref)}
                server={_this.api_upload_address}
                onprocessfile={() => this.props.getHandler()}
                onerror={error => this.displayError(error)}
              />
            </Col>
          </Row>
        </Container>
      </div>
    );
  }
}

export default CsvUpload;
