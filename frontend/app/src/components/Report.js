import React, { Component } from "react";
import { Container, Row, Col, Badge } from "react-bootstrap";
import axios from "axios";
import CsvUpload from "./CsvUpload";
import GetReport from "./GetReport";

class Report extends Component {
  constructor(props) {
    super(props);
    this.state = {
      contents: []
    };
    this.getContent = this.getContent.bind(this);
  }
  getContent() {
    let _this = this;
    _this.api_getReport_address = "http://localhost:80/getReport";
    axios.get(_this.api_getReport_address).then(function(response) {
      _this.setState({ contents: response.data });
    });
  }
  render() {
    return (
      <div>
        <header className="App-header ">
          <Container>
            <Row>
              <Col />
              <Col>
                <h1>
                  <Badge variant="secondary">Payroll Web Application</Badge>
                </h1>
              </Col>
              <Col />
            </Row>
            <Row>
              <Col />
              <Col xs={10}>
                <CsvUpload
                  getHandler={this.getContent}
                  contents={this.state.contents}
                />
              </Col>
              <Col />
            </Row>
            <Row>
              <Col />
              <Col xs={10}>
                <GetReport
                  getHandler={this.getContent}
                  contents={this.state.contents}
                />
              </Col>
              <Col />
            </Row>
          </Container>
        </header>
      </div>
    );
  }
}

export default Report;
