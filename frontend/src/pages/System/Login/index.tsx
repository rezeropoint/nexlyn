import { login } from "@/services/user";
import { LockOutlined, MobileOutlined, UserOutlined } from "@ant-design/icons";
import {
  LoginFormPage,
  ProConfigProvider,
  ProFormCaptcha,
  ProFormCheckbox,
  ProFormText,
} from "@ant-design/pro-components";
import {
  FormattedMessage,
  Helmet,
  SelectLang,
  useIntl,
  useModel,
} from "@umijs/max";
import { Alert, App, Tabs, theme } from "antd";
import { createStyles } from "antd-style";
import React, { useState } from "react";
import { flushSync } from "react-dom";
import Settings from "../../../../config/defaultSettings";
import styles from "./index.less";

const useStyles = createStyles(({ token }) => {
  return {
    lang: {
      width: 42,
      height: 42,
      lineHeight: "42px",
      position: "fixed",
      right: 16,
      top: 16,
      borderRadius: token.borderRadius,
      ":hover": {
        backgroundColor: token.colorBgTextHover,
      },
    },
  };
});

const Lang = () => {
  const { styles } = useStyles();

  return (
    <div className={styles.lang} data-lang>
      {SelectLang && <SelectLang />}
    </div>
  );
};

const LoginMessage: React.FC<{
  content: string;
}> = ({ content }) => {
  return (
    <Alert
      className={styles.loginMessage}
      message={content}
      type="error"
      showIcon
    />
  );
};

const LoginPage: React.FC = () => {
  const [userLoginState, setUserLoginState] = useState<API.LoginResponse>({
    code: 0,
  });
  const [type, setType] = useState<string>("account");
  const { initialState, setInitialState } = useModel("@@initialState");
  const { message } = App.useApp();
  const intl = useIntl();
  const { token } = theme.useToken();

  const fetchUserInfo = async () => {
    const userInfo = await initialState?.fetchUserInfo?.();
    if (userInfo) {
      flushSync(() => {
        setInitialState((s) => ({
          ...s,
          currentUser: userInfo,
        }));
      });
    }
  };

  const handleSubmit = async (values: API.LoginParams) => {
    try {
      // 登录
      const msg = await login({ ...values, type });
      if (msg.code === 0 && msg.token) {
        const defaultLoginSuccessMessage = intl.formatMessage({
          id: "pages.login.success",
          defaultMessage: "登录成功！",
        });
        message.success(msg.msg || defaultLoginSuccessMessage);
        // 保存 token 到本地存储
        localStorage.setItem("token", msg.token);
        await fetchUserInfo();
        const urlParams = new URL(window.location.href).searchParams;
        window.location.href = urlParams.get("redirect") || "/";
        return;
      }
      // 如果失败，显示后端返回的错误信息或通用错误信息
      const errorMessage =
        msg.msg ||
        intl.formatMessage({
          id: "pages.login.failure",
          defaultMessage: "登录失败，请重试！",
        });
      message.error(errorMessage);
      setUserLoginState({ ...msg, code: msg.code || -1 }); // 更新状态以显示错误
    } catch (error) {
      const defaultLoginFailureMessage = intl.formatMessage({
        id: "pages.login.failure",
        defaultMessage: "登录失败，请重试！",
      });
      console.log(error);
      message.error(defaultLoginFailureMessage);
    }
  };

  const { code } = userLoginState;

  return (
    <div className={styles.pageContainer}>
      <Helmet>
        <title>
          {intl.formatMessage({
            id: "menu.login",
            defaultMessage: "登录页",
          })}
          {Settings.title && ` - ${Settings.title}`}
        </title>
      </Helmet>
      <Lang />
      <LoginFormPage
        backgroundImageUrl="https://mdn.alipayobjects.com/huamei_gcee1x/afts/img/A*y0ZTS6WLwvgAAAAAAAAAAAAADml6AQ/fmt.webp"
        logo={<img alt="logo" src="/logos/icon.svg" />}
        backgroundVideoUrl="https://gw.alipayobjects.com/v/huamei_gcee1x/afts/video/jXRBRK_VAwoAAAAAAAAAAAAAK4eUAQBr"
        title="Nexlyn"
        containerStyle={{
          backgroundColor: token.colorBgMask,
          backdropFilter: "blur(4px)",
        }}
        subTitle={intl.formatMessage({
          id: "pages.layouts.userLayout.title",
          defaultMessage: "全球领先的智慧物联管理平台",
        })}
        onFinish={async (values) => {
          await handleSubmit(values as API.LoginParams);
        }}
      >
        <Tabs
          centered
          activeKey={type}
          onChange={setType}
          items={[
            {
              key: "account",
              label: intl.formatMessage({
                id: "pages.login.accountLogin.tab",
                defaultMessage: "账户密码登录",
              }),
            },
            {
              key: "mobile",
              label: intl.formatMessage({
                id: "pages.login.phoneLogin.tab",
                defaultMessage: "手机号登录",
              }),
              disabled: true,
            },
          ]}
        />

        {code !== 0 && type === "account" && (
          <LoginMessage
            content={
              userLoginState.msg ||
              intl.formatMessage({
                id: "pages.login.accountLogin.errorMessage",
                defaultMessage: "账户或密码错误",
              })
            }
          />
        )}

        {type === "account" && (
          <>
            <ProFormText
              name="userName"
              fieldProps={{
                size: "large",
                prefix: <UserOutlined className={styles.formIcon} />,
              }}
              placeholder={intl.formatMessage({
                id: "pages.login.username.placeholder",
                defaultMessage: "用户名",
              })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.username.required"
                      defaultMessage="请输入用户名!"
                    />
                  ),
                },
              ]}
            />
            <ProFormText.Password
              name="password"
              fieldProps={{
                size: "large",
                prefix: <LockOutlined className={styles.formIcon} />,
              }}
              placeholder={intl.formatMessage({
                id: "pages.login.password.placeholder",
                defaultMessage: "密码",
              })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.password.required"
                      defaultMessage="请输入密码！"
                    />
                  ),
                },
              ]}
            />
          </>
        )}

        {type === "mobile" && (
          <>
            <ProFormText
              fieldProps={{
                size: "large",
                prefix: <MobileOutlined className={styles.formIcon} />,
                disabled: true,
              }}
              name="mobile"
              placeholder={intl.formatMessage({
                id: "pages.login.phoneNumber.placeholder",
                defaultMessage: "请输入手机号（暂不可用）",
              })}
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.phoneNumber.required"
                      defaultMessage="请输入手机号！"
                    />
                  ),
                },
                {
                  pattern: /^1\d{10}$/,
                  message: (
                    <FormattedMessage
                      id="pages.login.phoneNumber.invalid"
                      defaultMessage="手机号格式错误！"
                    />
                  ),
                },
              ]}
            />
            <ProFormCaptcha
              fieldProps={{
                size: "large",
                prefix: <LockOutlined className={styles.formIcon} />,
                disabled: true,
              }}
              captchaProps={{
                size: "large",
                disabled: true,
              }}
              placeholder={intl.formatMessage({
                id: "pages.login.captcha.placeholder",
                defaultMessage: "请输入验证码（暂不可用）",
              })}
              captchaTextRender={(_timing, _count) => {
                return intl.formatMessage({
                  id: "pages.login.phoneLogin.getVerificationCode",
                  defaultMessage: "获取验证码（暂不可用）",
                });
              }}
              name="captcha"
              rules={[
                {
                  required: true,
                  message: (
                    <FormattedMessage
                      id="pages.login.captcha.required"
                      defaultMessage="请输入验证码！"
                    />
                  ),
                },
              ]}
              onGetCaptcha={async (_phone) => {
                message.info("手机号登录功能暂未开放");
              }}
            />
          </>
        )}

        <div className={styles.rememberContainer}>
          <ProFormCheckbox noStyle name="autoLogin">
            <span style={{ color: token.colorText }}>
              <FormattedMessage
                id="pages.login.rememberMe"
                defaultMessage="自动登录"
              />
            </span>
          </ProFormCheckbox>
          <a
            className={styles.forgotPasswordLink}
            style={{ color: token.colorPrimary }}
          >
            <FormattedMessage
              id="pages.login.forgotPassword"
              defaultMessage="忘记密码"
            />
          </a>
        </div>
      </LoginFormPage>
    </div>
  );
};

const Login: React.FC = () => {
  return (
    <ProConfigProvider dark>
      <LoginPage />
    </ProConfigProvider>
  );
};

export default Login;
